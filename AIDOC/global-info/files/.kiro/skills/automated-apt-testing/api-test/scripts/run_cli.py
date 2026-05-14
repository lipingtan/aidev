import os
import json
import sys
from pathlib import Path
from urllib.parse import urlparse, urljoin
from urllib.request import Request, urlopen
from urllib.error import URLError, HTTPError
import concurrent.futures

SKILL_DIR = Path(__file__).parent.parent
CONFIG_FILE = SKILL_DIR / 'env.config.json'
REPORT_DIR = SKILL_DIR / 'test-reports'

def load_config():
    if not CONFIG_FILE.exists():
        print(f'配置文件不存在: {CONFIG_FILE}')
        return None
    
    try:
        with open(CONFIG_FILE, 'r', encoding='utf-8') as f:
            return json.load(f)
    except json.JSONDecodeError as e:
        print(f'解析配置文件失败: {e}')
        return None

def get_environment_url(config, env_name):
    if not config or 'environments' not in config or env_name not in config['environments']:
        print(f'环境 {env_name} 未在配置文件中定义')
        return None
    return config['environments'][env_name]['baseUrl']

def http_request(options):
    url = options['url']
    method = options.get('method', 'GET')
    headers = options.get('headers', {})
    body = options.get('body')
    
    req = Request(url, method=method)
    
    for key, value in headers.items():
        req.add_header(key, value)
    
    if body:
        body_str = json.dumps(body) if isinstance(body, dict) else body
        req.add_header('Content-Type', 'application/json')
        req.data = body_str.encode('utf-8')
    
    try:
        with urlopen(req) as response:
            data = response.read().decode('utf-8')
            try:
                body = json.loads(data) if data else {}
            except json.JSONDecodeError:
                body = {'raw': data}
            
            return {
                'statusCode': response.status,
                'headers': dict(response.headers),
                'body': body
            }
    except HTTPError as e:
        return {
            'statusCode': e.code,
            'headers': dict(e.headers),
            'body': {'error': str(e)}
        }
    except URLError as e:
        raise Exception(f'请求失败: {e}')

def run_test(test_case, base_url):
    full_url = urljoin(base_url, test_case['request']['url'])
    
    print(f'\n执行测试: {test_case["name"]}')
    print(f'请求: {test_case["request"]["method"]} {full_url}')
    
    if 'params' in test_case['request']:
        from urllib.parse import urlencode, urlparse
        parsed = urlparse(full_url)
        query = urlencode(test_case['request']['params'])
        full_url = parsed._replace(query=query).geturl()
    
    if 'body' in test_case['request']:
        print(f'请求体: {test_case["request"]["body"]}')
    
    try:
        response = http_request({
            'url': full_url,
            'method': test_case['request']['method'],
            'headers': test_case['request'].get('headers', {}),
            'body': test_case['request'].get('body')
        })
        
        result = {
            'name': test_case['name'],
            'success': True,
            'request': test_case['request'],
            'response': response,
            'expected': test_case['expect']
        }
        
        status_code_match = response['statusCode'] == test_case['expect']['statusCode']
        body_match = True
        
        if 'body' in test_case['expect']:
            expected_body = test_case['expect']['body']
            actual_body = response['body']
            
            if isinstance(expected_body, dict) and isinstance(actual_body, dict):
                body_match = all(
                    key in actual_body and actual_body[key] == expected_body[key]
                    for key in expected_body
                )
            else:
                body_match = json.dumps(response['body'], sort_keys=True) == json.dumps(test_case['expect']['body'], sort_keys=True)
        
        result['passed'] = status_code_match and body_match
        result['details'] = {
            'statusCodeMatch': status_code_match,
            'bodyMatch': body_match,
            'expectedStatusCode': test_case['expect']['statusCode'],
            'actualStatusCode': response['statusCode'],
            'expectedBody': test_case['expect'].get('body'),
            'actualBody': response['body']
        }
        
        return result
    except Exception as error:
        return {
            'name': test_case['name'],
            'success': False,
            'error': str(error)
        }

def run_test_file(test_file, env_name):
    config = load_config()
    base_url = get_environment_url(config, env_name)
    
    if not base_url:
        raise Exception('无法获取环境配置')
    
    test_file_path = SKILL_DIR / test_file
    if not test_file_path.exists():
        raise Exception(f'测试文件不存在: {test_file_path}')
    
    test_file_name = test_file_path.name
    
    REPORT_DIR.mkdir(parents=True, exist_ok=True)
    
    print('\n========================================')
    print(f'执行测试: {test_file_name}')
    print(f'环境: {env_name} ({base_url})')
    print('========================================')
    
    try:
        with open(test_file_path, 'r', encoding='utf-8') as f:
            test_suite = json.load(f)
        
        results = []
        for test_case in test_suite['tests']:
            result = run_test(test_case, base_url)
            results.append(result)
        
        report = {
            'suiteName': test_suite['name'],
            'environment': env_name,
            'baseUrl': base_url,
            'timestamp': json.dumps({}).replace('{}', ''),  # ISO format
            'stats': {
                'total': len(results),
                'passed': sum(1 for r in results if r.get('passed')),
                'failed': sum(1 for r in results if not r.get('passed'))
            },
            'results': results
        }
        
        report_file = REPORT_DIR / f'{test_file_path.stem}-{env_name}.json'
        with open(report_file, 'w', encoding='utf-8') as f:
            json.dump(report, f, indent=2, ensure_ascii=False)
        
        text_report_file = REPORT_DIR / f'{test_file_path.stem}-{env_name}.txt'
        generate_text_report(report, text_report_file)
        
        return {'success': True, 'report': report, 'testFile': test_file, 'envName': env_name}
    except Exception as error:
        raise error

def generate_text_report(report, report_file):
    with open(report_file, 'w', encoding='utf-8') as f:
        f.write('=' * 80 + '\n')
        f.write(f'测试报告: {report["suiteName"]}\n')
        f.write('=' * 80 + '\n\n')
        
        f.write(f'环境: {report["environment"]}\n')
        f.write(f'基础URL: {report["baseUrl"]}\n')
        f.write(f'执行时间: {report["timestamp"] or "N/A"}\n\n')
        
        f.write('-' * 80 + '\n')
        f.write('测试统计\n')
        f.write('-' * 80 + '\n')
        stats = report['stats']
        f.write(f'总用例数: {stats["total"]}\n')
        f.write(f'通过: {stats["passed"]}\n')
        f.write(f'失败: {stats["failed"]}\n')
        if stats['total'] > 0:
            pass_rate = (stats['passed'] / stats['total'] * 100)
            f.write(f'通过率: {pass_rate:.2f}%\n')
        f.write('\n')
        
        f.write('-' * 80 + '\n')
        f.write('测试详情\n')
        f.write('-' * 80 + '\n\n')
        
        for idx, result in enumerate(report['results'], 1):
            status = '✅ 通过' if result.get('passed') else '❌ 失败'
            f.write(f'{idx}. {result["name"]}\n')
            f.write(f'   状态: {status}\n')
            
            if 'request' in result:
                req = result['request']
                f.write(f'   请求: {req["method"]} {req["url"]}\n')
                if 'params' in req:
                    f.write(f'   参数: {req["params"]}\n')
                if 'body' in req:
                    f.write(f'   请求体: {req["body"]}\n')
            
            if 'response' in result:
                resp = result['response']
                f.write(f'   响应状态码: {resp["statusCode"]}\n')
                if 'body' in resp:
                    f.write(f'   响应体: {json.dumps(resp["body"], ensure_ascii=False)}\n')
            
            if not result.get('passed') and 'details' in result:
                details = result['details']
                f.write(f'   失败原因:\n')
                if not details.get('statusCodeMatch'):
                    f.write(f'     - 状态码不匹配: 预期 {details["expectedStatusCode"]}, 实际 {details["actualStatusCode"]}\n')
                if not details.get('bodyMatch'):
                    f.write(f'     - 响应体不匹配\n')
                    f.write(f'       预期: {json.dumps(details["expectedBody"], ensure_ascii=False)}\n')
                    f.write(f'       实际: {json.dumps(details["actualBody"], ensure_ascii=False)}\n')
            
            f.write('\n')
        
        f.write('=' * 80 + '\n')
        f.write('报告结束\n')
        f.write('=' * 80 + '\n')

def parse_results(results):
    print('\n========================================')
    print('测试结果汇总')
    print('========================================\n')
    
    total_tests = 0
    total_passed = 0
    total_failed = 0
    
    for index, result in enumerate(results, 1):
        print(f'\n{index}. {Path(result["testFile"]).name} ({result["envName"]})')
        
        if 'report' in result:
            stats = result['report']['stats']
            total_tests += stats['total']
            total_passed += stats['passed']
            total_failed += stats['failed']
            
            print(f'   总数: {stats["total"]}')
            print(f'   通过: {stats["passed"]}')
            print(f'   失败: {stats["failed"]}')
            
            if stats['total'] > 0:
                pass_rate = (stats['passed'] / stats['total'] * 100)
                print(f'   通过率: {pass_rate:.2f}%')
            
            failed_tests = [r for r in result['report']['results'] if not r.get('passed')]
            if failed_tests:
                print('\n   失败用例:')
                for i, test in enumerate(failed_tests, 1):
                    print(f'   {i}. {test["name"]}')
                    print(f'      预期状态码: {test["details"]["expectedStatusCode"]}')
                    print(f'      实际状态码: {test["details"]["actualStatusCode"]}')
                    if not test['details']['statusCodeMatch']:
                        print(f'      状态码不匹配')
                    if not test['details']['bodyMatch']:
                        print(f'      响应体不匹配')
        else:
            print('   执行完成，无详细报告')
    
    print('\n========================================')
    print('总体统计')
    print('========================================')
    print(f'总用例数: {total_tests}')
    print(f'总通过: {total_passed}')
    print(f'总失败: {total_failed}')
    
    if total_tests > 0:
        overall_pass_rate = (total_passed / total_tests * 100)
        print(f'总体通过率: {overall_pass_rate:.2f}%')
    
    print(f'\n报告目录: {REPORT_DIR}')
    print(f'JSON 报告: {REPORT_DIR / "*.json"}')
    print(f'文本报告: {REPORT_DIR / "*.txt"}')

def run_tests(test_files, env_name, parallel=False):
    print(f'\n执行模式: {"并行执行" if parallel else "顺序执行"}')
    print(f'测试数量: {len(test_files)}')
    print(f'目标环境: {env_name}\n')
    
    results = []
    
    if parallel:
        with concurrent.futures.ThreadPoolExecutor() as executor:
            futures = {
                executor.submit(run_test_file, file, env_name): file
                for file in test_files
            }
            
            for future in concurrent.futures.as_completed(futures):
                file = futures[future]
                try:
                    result = future.result()
                    results.append(result)
                except Exception as error:
                    results.append({
                        'success': False,
                        'error': str(error),
                        'testFile': file,
                        'envName': env_name
                    })
    else:
        for file in test_files:
            try:
                result = run_test_file(file, env_name)
                results.append(result)
            except Exception as error:
                results.append({
                    'success': False,
                    'error': str(error),
                    'testFile': file,
                    'envName': env_name
                })
    
    parse_results(results)
    
    return 0 if all(r['success'] for r in results) else 1

if __name__ == '__main__':
    if len(sys.argv) < 3:
        print('用法: python run_cli.py <测试文件路径> <环境名> [并行]')
        print('示例: python run_cli.py tests/simple-controller.json dev')
        print('示例: python run_cli.py tests/*.json test parallel')
        sys.exit(1)
    
    test_file = sys.argv[1]
    env_name = sys.argv[2]
    parallel = 'parallel' in sys.argv
    
    exit_code = run_tests([test_file], env_name, parallel)
    sys.exit(exit_code)
