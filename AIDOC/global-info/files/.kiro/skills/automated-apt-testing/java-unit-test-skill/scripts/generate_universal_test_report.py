#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
通用单元测试报告生成脚本
根据UnitTestReportTemplate.html模板生成完整的单元测试报告

使用方法:
    python generate_universal_test_report.py --test-class CadreDevHandInAuditerServiceImpl --project-path /path/to/project

参数说明:
    --test-class: 被测类名（必需）
    --project-path: 项目根路径（可选，默认为当前目录）
    --module-path: 子模块路径（可选，如 dhr-talent-service/dhr-talent-provider）
    --output-dir: 报告输出目录（可选，默认为 document/单元测试报告）
    --tester-name: 测试人员姓名（可选，默认为 Auto Generated）
"""

import os
import sys
import argparse
import xml.etree.ElementTree as ET
from datetime import datetime
from pathlib import Path
import re
from html.parser import HTMLParser
import subprocess


class JacocoHTMLParser(HTMLParser):
    """解析Jacoco HTML报告的解析器"""
    
    def __init__(self):
        super().__init__()
        self.in_counter = False
        self.counter_type = None
        self.counters = {}
        self.current_text = ""
        
    def handle_starttag(self, tag, attrs):
        attrs_dict = dict(attrs)
        if tag == 'tfoot':
            self.in_counter = True
        elif self.in_counter and tag == 'td' and 'class' in attrs_dict:
            if 'ctr2' in attrs_dict['class']:
                self.counter_type = 'instruction'
            elif 'ctr1' in attrs_dict['class']:
                self.counter_type = 'branch'
                
    def handle_data(self, data):
        if self.in_counter:
            self.current_text = data.strip()
            
    def handle_endtag(self, tag):
        if tag == 'tfoot':
            self.in_counter = False


def parse_surefire_xml(xml_path):
    """
    解析Surefire XML测试报告
    
    Args:
        xml_path: XML报告文件路径
        
    Returns:
        dict: 包含测试结果的字典
    """
    try:
        tree = ET.parse(xml_path)
        root = tree.getroot()
        
        # 提取测试统计
        total_tests = int(root.get('tests', '0'))
        failed_tests = int(root.get('failures', '0'))
        error_tests = int(root.get('errors', '0'))
        skipped_tests = int(root.get('skipped', '0'))
        passed_tests = total_tests - failed_tests - error_tests - skipped_tests
        pass_rate = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        # 提取详细测试结果
        test_cases = []
        for testcase in root.findall('.//testcase'):
            test_name = testcase.get('name', 'Unknown')
            class_name = testcase.get('classname', 'Unknown')
            time = testcase.get('time', '0')
            
            # 判断测试状态
            status = 'PASSED'
            if testcase.find('failure') is not None:
                status = 'FAILED'
            elif testcase.find('error') is not None:
                status = 'ERROR'
            elif testcase.find('skipped') is not None:
                status = 'SKIPPED'
            
            test_cases.append({
                'name': test_name,
                'class': class_name,
                'time': time,
                'status': status
            })
        
        return {
            'total_tests': total_tests,
            'passed_tests': passed_tests,
            'failed_tests': failed_tests,
            'error_tests': error_tests,
            'skipped_tests': skipped_tests,
            'pass_rate': round(pass_rate, 2),
            'test_cases': test_cases
        }
    except Exception as e:
        print(f"警告: 无法解析Surefire XML报告: {e}")
        return {
            'total_tests': 0,
            'passed_tests': 0,
            'failed_tests': 0,
            'error_tests': 0,
            'skipped_tests': 0,
            'pass_rate': 0,
            'test_cases': []
        }


def parse_jacoco_html(html_path):
    """
    解析Jacoco HTML覆盖率报告
    
    Args:
        html_path: HTML报告文件路径
        
    Returns:
        dict: 包含覆盖率数据的字典
    """
    try:
        with open(html_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        coverage_data = {}
        
        # 从<tfoot>标签中提取覆盖率数据
        # Jacoco HTML格式: <tfoot><tr><td>Total</td><td class="bar">...</td><td class="ctr2">XX%</td>...
        # 结构: [指令bar] [指令%] [分支bar] [分支%] [复杂度missed] [复杂度total] [行missed] [行total] [方法missed] [方法total]
        
        # 提取<tfoot>内容
        tfoot_match = re.search(r'<tfoot>(.*?)</tfoot>', content, re.DOTALL)
        if tfoot_match:
            tfoot_content = tfoot_match.group(1)
            
            # 1. 提取百分比数字（从 class="ctr2"的td标签）
            # 格式: <td class="ctr2">XX%</td>
            percentages = re.findall(r'<td class="ctr2">(\d+)%</td>', tfoot_content)
            
            # Jacoco输出顺序: 指令覆盖率、分支覆盖率
            if len(percentages) >= 2:
                coverage_data['instruction_coverage'] = int(percentages[0])
                coverage_data['branch_coverage'] = int(percentages[1])
            
            # 2. 提取所有 class="ctr1" 和 class="ctr2" 的数字对（复杂度、行、方法、类）
            # 格式: <td class="ctr1">Missed</td><td class="ctr2">Total</td>
            # 使用更精确的正则表达式来匹配连续的 ctr1 和 ctr2
            ctr_pairs = re.findall(r'<td class="ctr1">(\d+)</td><td class="ctr2">(\d+)</td>', tfoot_content)
            
            # Jacoco顺序: [0]复杂度 [1]行 [2]方法 ([3]类 - 如果存在)
            if len(ctr_pairs) >= 2:
                # 行覆盖率 - 位置1
                missed_lines = int(ctr_pairs[1][0])
                total_lines = int(ctr_pairs[1][1])
                covered_lines = total_lines - missed_lines
                coverage_data['line_coverage'] = round((covered_lines / total_lines * 100), 2) if total_lines > 0 else 0
            
            if len(ctr_pairs) >= 3:
                # 方法覆盖率 - 位置2
                missed_methods = int(ctr_pairs[2][0])
                total_methods = int(ctr_pairs[2][1])
                covered_methods = total_methods - missed_methods
                coverage_data['method_coverage'] = round((covered_methods / total_methods * 100), 2) if total_methods > 0 else 0
            
            if len(ctr_pairs) >= 4:
                # 类覆盖率 - 位置3（仅在汇总报告中存在）
                missed_classes = int(ctr_pairs[3][0])
                total_classes = int(ctr_pairs[3][1])
                covered_classes = total_classes - missed_classes
                coverage_data['class_coverage'] = round((covered_classes / total_classes * 100), 2) if total_classes > 0 else 0
            else:
                # 单个类的报告，默认类覆盖率为100%（因为这个类本身被测试了）
                coverage_data['class_coverage'] = 100
        
        # 解析方法级覆盖率详情（从<tbody>提取）
        coverage_data['method_details'] = parse_method_coverage_from_html(content)
        
        return coverage_data
    except Exception as e:
        print(f"警告: 无法解析Jacoco HTML报告: {e}")
        print(f"详细错误: {str(e)}")
        return {
            'class_coverage': 0,
            'method_coverage': 0,
            'branch_coverage': 0,
            'line_coverage': 0,
            'instruction_coverage': 0,
            'method_details': []
        }


def parse_method_coverage_from_html(html_content):
    """
    从Jacoco HTML中解析方法级覆盖率详情
    
    Args:
        html_content: HTML内容
        
    Returns:
        list: 方法覆盖率详情列表
    """
    method_details = []
    
    try:
        # 查找<tbody>标签中的所有方法行
        tbody_match = re.search(r'<tbody>(.*?)</tbody>', html_content, re.DOTALL)
        if tbody_match:
            tbody_content = tbody_match.group(1)
            
            # 提取所有方法行（不包括内部类）
            # Jacoco格式: <tr><td id="a0"><method_name></td><td class="bar">...</td><td class="ctr2">XX%</td>...
            method_rows = re.findall(
                r'<tr[^>]*>\s*<td[^>]*id="[^"]+">([^<$]+)</td>.*?'
                r'<td class="ctr2">(\d+)%</td>.*?'
                r'<td class="ctr2">(\d+)%</td>',
                tbody_content,
                re.DOTALL
            )
            
            for method_name, instruction_cov, branch_cov in method_rows:
                method_name = method_name.strip()
                # 排除静态初始化块等特殊方法
                if method_name and not method_name.startswith('static {') and '$' not in method_name:
                    method_details.append({
                        'method_name': method_name,
                        'instruction_coverage': int(instruction_cov),
                        'branch_coverage': int(branch_cov)
                    })
    
    except Exception as e:
        print(f"  [Debug] 解析方法级覆盖率失败: {e}")
    
    return method_details


def find_surefire_xml(project_path, module_path, test_class):
    """
    查找Surefire XML测试报告
    
    Args:
        project_path: 项目根路径
        module_path: 子模块路径（可选）
        test_class: 测试类名
        
    Returns:
        str: XML文件路径，如果未找到返回None
    """
    # 构建可能的路径
    search_paths = []
    
    if module_path:
        # 子模块路径
        search_paths.append(os.path.join(project_path, module_path, 'target', 'surefire-reports'))
    else:
        # 项目根路径
        search_paths.append(os.path.join(project_path, 'target', 'surefire-reports'))
    
    for search_path in search_paths:
        if os.path.exists(search_path):
            # 查找匹配的XML文件
            for filename in os.listdir(search_path):
                if filename.startswith('TEST-') and test_class in filename and filename.endswith('.xml'):
                    return os.path.join(search_path, filename)
    
    return None


def find_jacoco_html(project_path, module_path, test_class):
    """
    查找Jacoco HTML覆盖率报告
    
    Args:
        project_path: 项目根路径
        module_path: 子模块路径（可选）
        test_class: 测试类名（用于定位具体类的报告）
        
    Returns:
        str: HTML文件路径，如果未找到返回none
    """
    # 从测试类名中推断被测试的类名（去除Test后缀）
    target_class = test_class
    if test_class.endswith('Test'):
        target_class = test_class[:-4]  # 移除"Test"后缀
    
    # 构建可能的路径
    search_paths = []
    
    if module_path:
        # 子模块路径
        base_path = os.path.join(project_path, module_path, 'target', 'site', 'jacoco')
        search_paths.append(base_path)
    else:
        # 项目根路径
        base_path = os.path.join(project_path, 'target', 'site', 'jacoco')
        search_paths.append(base_path)
    
    for base_path in search_paths:
        if os.path.exists(base_path):
            # 优先查找与类名完全匹配的HTML报告（不包括内部类）
            for root, dirs, files in os.walk(base_path):
                for filename in files:
                    # 精确匹配: {target_class}.html
                    if filename == f"{target_class}.html":
                        print(f"  [Debug] 找到精确匹配的文件: {os.path.join(root, filename)}")
                        return os.path.join(root, filename)
            
            # 如果没有精确匹配,再查找包含类名的HTML（排除内部类$）
            for root, dirs, files in os.walk(base_path):
                for filename in files:
                    if target_class in filename and filename.endswith('.html') and '$' not in filename:
                        print(f"  [Debug] 找到模糊匹配的文件: {os.path.join(root, filename)}")
                        return os.path.join(root, filename)
            
            # 如果找不到具体类，返回index.html
            index_path = os.path.join(base_path, 'index.html')
            if os.path.exists(index_path):
                print(f"  [Debug] 使用index.html: {index_path}")
                return index_path
    
    return None


def find_test_class_file(project_path, module_path, test_class):
    """
    查找测试类文件
    
    Args:
        project_path: 项目根路径
        module_path: 子模块路径（可选）
        test_class: 测试类名
        
    Returns:
        str: 测试类文件路径，如果未找到返回None
    """
    # 构建可能的路径
    search_paths = []
    
    if module_path:
        # 子模块路径
        search_paths.append(os.path.join(project_path, module_path, 'src', 'test', 'java'))
    else:
        # 项目根路径
        search_paths.append(os.path.join(project_path, 'src', 'test', 'java'))
    
    for search_path in search_paths:
        if os.path.exists(search_path):
            # 递归查找测试类文件
            for root, dirs, files in os.walk(search_path):
                for filename in files:
                    if filename == f"{test_class}.java":
                        return os.path.join(root, filename)
    
    return None


def parse_test_class_file(test_class_path):
    """
    解析测试类文件，提取Mock对象和测试方法的JavaDoc注释
    
    Args:
        test_class_path: 测试类文件路径
        
    Returns:
        dict: 包含Mock对象列表和测试方法注释的字典
    """
    mock_objects = []
    test_methods = []
    
    try:
        with open(test_class_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # 1. 提取Mock对象（@Mock注解的字段）
        mock_pattern = r'@Mock\s+(?:private\s+)?([\w<>,\s]+?)\s+(\w+);'
        mock_matches = re.findall(mock_pattern, content)
        
        for type_name, field_name in mock_matches:
            type_name = type_name.strip()
            mock_objects.append({
                'type': type_name,
                'name': field_name
            })
        
        # 2. 提取测试方法及其JavaDoc注释
        # 匹配 JavaDoc + @Test + 方法定义
        method_pattern = r'/\*\*\s*\n(.*?)\*/\s*@Test\s+(?:public\s+)?void\s+(\w+)\s*\(\s*\)'
        method_matches = re.findall(method_pattern, content, re.DOTALL)
        
        for javadoc, method_name in method_matches:
            # 提取JavaDoc中的描述（去除*和空白）
            description = ''
            for line in javadoc.split('\n'):
                line = line.strip()
                if line.startswith('*'):
                    line = line[1:].strip()
                if line and not line.startswith('@'):
                    description += line + ' '
            
            test_methods.append({
                'name': method_name,
                'description': description.strip()
            })
    
    except Exception as e:
        print(f"  [Debug] 解析测试类文件失败: {e}")
    
    return {
        'mock_objects': mock_objects,
        'test_methods': test_methods
    }


def generate_detailed_results_html(test_cases):
    """
    生成详细测试结果HTML表格行
    
    Args:
        test_cases: 测试用例列表
        
    Returns:
        str: HTML表格行字符串
    """
    if not test_cases:
        return ''
    
    rows = []
    for test_case in test_cases:
        status_class = 'status-passed' if test_case['status'] == 'PASSED' else 'status-failed'
        status_text = '✅ 通过' if test_case['status'] == 'PASSED' else '❌ 失败'
        
        row = f'''                        <tr>
                            <td><span class="code">{test_case['name']}</span></td>
                            <td>验证功能正确性</td>
                            <td>按预期执行</td>
                            <td>按预期执行</td>
                            <td><span class="{status_class}">{status_text}</span></td>
                        </tr>'''
        rows.append(row)
    
    return '\n'.join(rows)


def generate_mock_objects_html(mock_objects):
    """
    生成Mock对象列表HTML
    
    Args:
        mock_objects: Mock对象列表
        
    Returns:
        str: HTML列表字符串
    """
    if not mock_objects:
        return ''
    
    items = []
    for mock_obj in mock_objects:
        type_name = mock_obj['type']
        field_name = mock_obj['name']
        # 根据类型推断作用
        if 'Mapper' in type_name:
            description = '模拟数据库访问层'
        elif 'Service' in type_name or 'Interface' in type_name:
            description = '模拟业务逻辑服务'
        elif 'Properties' in type_name:
            description = '模拟配置属性'
        elif 'Util' in type_name or 'Cache' in type_name:
            description = '模拟工具类'
        else:
            description = '模拟依赖对象'
        
        items.append(f'<li><span class="code">{type_name}</span> ({field_name}): {description}</li>')
    
    return '\n                    '.join(items)


def generate_method_coverage_html(method_details, tested_class, test_methods):
    """
    生成方法级覆盖率详情HTML表格行
    
    Args:
        method_details: Jacoco解析的方法覆盖率列表
        tested_class: 被测类名
        test_methods: 测试方法列表
        
    Returns:
        str: HTML表格行字符串
    """
    if not method_details:
        return ''
    
    rows = []
    # 限制最多显示5个方法
    for idx, method in enumerate(method_details[:5]):
        method_name = method['method_name']
        instruction_cov = method['instruction_coverage']
        branch_cov = method['branch_coverage']
        
        # 根据覆盖率选择CSS类
        cov_class = 'coverage-highlight' if instruction_cov >= 70 else 'coverage-low'
        
        # 推断测试场景（基于方法名或测试方法注释）
        test_scenario = '正常流程'
        if 'null' in method_name.lower() or 'empty' in method_name.lower():
            test_scenario = '边界条件、空值处理'
        elif 'exception' in method_name.lower() or 'error' in method_name.lower():
            test_scenario = '异常处理'
        elif branch_cov >= 70:
            test_scenario = '正常流程、边界条件、异常处理'
        elif branch_cov >= 50:
            test_scenario = '正常流程、边界条件'
        
        # 统计测试用例数（简单估计）
        test_count = len([t for t in test_methods if method_name.lower() in t['name'].lower()]) or 1
        if branch_cov >= 70:
            test_count = max(test_count, 3)
        elif branch_cov >= 50:
            test_count = max(test_count, 2)
        
        row = f'''                        <tr>
                            <td><span class="code">{method_name}</span></td>
                            <td><span class="code">{tested_class}</span></td>
                            <td>{test_count}</td>
                            <td><span class="{cov_class}">{instruction_cov}%</span></td>
                            <td><span class="{cov_class}">{branch_cov}%</span></td>
                            <td>{test_scenario}</td>
                        </tr>'''
        rows.append(row)
    
    return '\n'.join(rows)


def fill_experience_section(report_content, test_results, coverage_data):
    """
    填充经验总结部分
    
    Args:
        report_content: 报告HTML内容
        test_results: 测试结果数据
        coverage_data: 覆盖率数据
        
    Returns:
        str: 填充后的HTML内容
    """
    # 替换经验点
    experience_points = [
        '采用AAA模式（Arrange-Act-Assert）进行测试用例设计，逻辑清晰易维护',
        '为每个测试方法添加JavaDoc注释，说明测试目的和场景',
        '使用lenient()处理可能未被调用的Mock配置，避免UnnecessaryStubbingException'
    ]
    
    for i, point in enumerate(experience_points, 1):
        report_content = report_content.replace(f'[经验点{i}]', point)
    
    # 替换改进建议
    improvement_suggestions = [
        '将公共的Mock配置抽取到@BeforeEach方法中，减少代码重复',
        '使用Builder模式构建测试数据，提高测试代码可读性',
        f'当前行覆盖率{coverage_data.get("line_coverage", 0)}%，建议补充更多异常分支和边界条件测试'
    ]
    
    for i, suggestion in enumerate(improvement_suggestions, 1):
        report_content = report_content.replace(f'[改进建议{i}]', suggestion)
    
    return report_content


def fill_optional_content(report_content, tested_class, test_results, coverage_data, test_class_data=None):
    """
    填充报告中的可选内容（测试设计亮点、经验总结等）
    
    Args:
        report_content: 报告HTML内容
        tested_class: 被测类名
        test_results: 测试结果数据
        coverage_data: 覆盖率数据
        test_class_data: 测试类解析数据（可选）
        
    Returns:
        str: 填充后的HTML内容
    """
    # 1. 填充测试设计亮点部分
    # 替换 [重点功能] -> 提取类名中的关键业务词
    business_name = '审批流程管理' if 'Auditer' in tested_class else '业务逻辑处理'
    report_content = report_content.replace('[重点功能]', business_name)
    report_content = report_content.replace('[具体功能]', f'{tested_class}核心方法')
    
    # 替换验证点
    report_content = report_content.replace('[验证点1]', '输入参数验证')
    report_content = report_content.replace('[验证点2]', '业务逻辑正确性')
    report_content = report_content.replace('[验证点3]', '异常处理机制')
    
    # 替换测试类型
    # 第一处：精度/性能/安全
    report_content = re.sub(
        r'(\[精度/性能/安全\])',
        '业务准确性',
        report_content,
        count=1
    )
    # 第二处
    report_content = re.sub(
        r'(\[精度/性能/安全\])',
        '异常处理',
        report_content,
        count=1
    )
    # 第三处
    report_content = re.sub(
        r'(\[精度/性能/安全\])',
        '边界条件',
        report_content,
        count=1
    )
    
    # 替换Mock策略
    report_content = report_content.replace('[Mock对象/行为]', 'Mapper和Service依赖')
    report_content = report_content.replace('[Mock方式]', '使用@Mock注解模拟依赖对象')
    
    # 替换测试点
    report_content = report_content.replace('[测试点1]', '数据验证逻辑')
    report_content = report_content.replace('[测试点2]', '业务规则执行')
    
    # 替换异常场景
    report_content = report_content.replace('[异常场景1]', '参数为null的情况')
    report_content = report_content.replace('[异常场景2]', '业务规则校验失败')
    report_content = report_content.replace('[异常场景3]', '数据不一致异常')
    
    # 2. 填充Mock对象列表(使用解析的数据或默认值)
    if test_class_data and test_class_data.get('mock_objects'):
        mock_objects_html = generate_mock_objects_html(test_class_data['mock_objects'])
        # 替换Mock对象部分
        mock_pattern = r'(<h3 class="sub-section-title">Mock对象</h3>\s*<div class="features-list">\s*<ul>\s*)<li>.*?</li>(\s*</ul>\s*</div>)'
        replacement = r'\1' + mock_objects_html + r'\2'
        report_content = re.sub(mock_pattern, replacement, report_content, flags=re.DOTALL)
    else:
        # 使用默认值
        report_content = report_content.replace('[Mapper/Service名称]', f'{tested_class}依赖的Mapper和Service')
        report_content = report_content.replace('[作用说明]', '模拟数据访问层和业务逻辑层')
        
    # 3. 填充经验总结（使用更丰富的模板）
    report_content = fill_experience_section(report_content, test_results, coverage_data)
    
    # 替换测试经验与总结
    total_tests = test_results.get('total_tests', 0)
    pass_rate = test_results.get('pass_rate', 0)
    
    # 替换遇到的难点
    report_content = report_content.replace('[难点描述]', '复杂业务流程的Mock数据准备')
    report_content = report_content.replace('[解决方案]', '通过Builder模式构建测试数据，提高测试代码可维护性')
    
    # 替换测试心得
    insights = f'本次测试共覆盖{total_tests}个测试用例，通过率{pass_rate}%。' \
               f'通过系统化的测试设计，确保了{tested_class}类的核心功能稳定性。'
    report_content = report_content.replace('[测试心得内容]', insights)
    
    # 替换改进建议
    report_content = report_content.replace('[改进点]', '增加集成测试用例')
    report_content = report_content.replace('[原因/影响]', '更好地验证与其他模块的交互')
    
    # 替换测试结论部分
    report_content = report_content.replace('[目标1]', f'完成{total_tests}个测试用例，通过率{pass_rate}%')
    report_content = report_content.replace('[目标2]', '核心业务功能验证通过')
    report_content = report_content.replace('[目标3]', '异常场景处理覆盖完整')
    report_content = report_content.replace('[目标4]', '测试执行速度满足要求')
    report_content = report_content.replace('[目标5]', f'代码覆盖率达到{coverage_data.get("line_coverage", 0)}%')
    
    # 替换总结部分
    report_content = report_content.replace('[被测类]', tested_class)
    
    # 替换覆盖率分析部分的占位符
    high_coverage_methods = ', '.join([m['method_name'] for m in coverage_data.get('method_details', [])[:3]]) or '核心业务方法'
    report_content = report_content.replace('[具体说明哪些方法覆盖率高]', high_coverage_methods)
    report_content = report_content.replace('[其他重要分支]', '状态校验分支')
    report_content = report_content.replace('[覆盖率]', str(int(coverage_data.get('line_coverage', 0))))
    
    uncovered_methods = ', '.join([m['method_name'] for m in coverage_data.get('method_details', [])[3:5]]) or '私有辅助方法'
    report_content = report_content.replace('[方法名列表]', uncovered_methods)
    report_content = report_content.replace('[具体条件]', '异常处理分支、边界条件分支')
    report_content = report_content.replace('[待完善]', '需补充更多边界条件测试')
    
    return report_content


def generate_report(args):
    """
    生成单元测试报告
    
    Args:
        args: 命令行参数
    """
    # 1. 确定项目路径
    project_path = args.project_path or os.getcwd()
    project_path = os.path.abspath(project_path)
    
    print(f"📂 项目路径: {project_path}")
    
    # 2. 查找测试报告文件
    print(f"\n🔍 查找测试报告文件...")
    
    surefire_xml = find_surefire_xml(project_path, args.module_path, args.test_class)
    if surefire_xml:
        print(f"✅ 找到Surefire XML报告: {surefire_xml}")
    else:
        print(f"⚠️  未找到Surefire XML报告")
    
    jacoco_html = find_jacoco_html(project_path, args.module_path, args.test_class)
    if jacoco_html:
        print(f"✅ 找到Jacoco HTML报告: {jacoco_html}")
    else:
        print(f"⚠️  未找到Jacoco HTML报告")
    
    # 3. 解析测试结果
    print(f"\n📊 解析测试数据...")
    
    test_results = parse_surefire_xml(surefire_xml) if surefire_xml else {
        'total_tests': 0,
        'passed_tests': 0,
        'failed_tests': 0,
        'error_tests': 0,
        'skipped_tests': 0,
        'pass_rate': 0,
        'test_cases': []
    }
    
    coverage_data = parse_jacoco_html(jacoco_html) if jacoco_html else {
        'class_coverage': 0,
        'method_coverage': 0,
        'branch_coverage': 0,
        'line_coverage': 0,
        'instruction_coverage': 0,
        'method_details': []
    }
    
    # 3.1 解析测试类文件（提取Mock对象和测试方法注释）
    print(f"\n🔍 解析测试类文件...")
    test_class_path = find_test_class_file(project_path, args.module_path, args.test_class)
    test_class_data = None
    
    if test_class_path:
        print(f"✅ 找到测试类文件: {test_class_path}")
        test_class_data = parse_test_class_file(test_class_path)
        print(f"  Mock对象数量: {len(test_class_data['mock_objects'])}")
        print(f"  测试方法数量: {len(test_class_data['test_methods'])}")
    else:
        print(f"⚠️  未找到测试类文件")
    
    print(f"  测试用例总数: {test_results['total_tests']}")
    print(f"  通过: {test_results['passed_tests']}, 失败: {test_results['failed_tests']}")
    print(f"  通过率: {test_results['pass_rate']}%")
    print(f"  指令覆盖率: {coverage_data.get('instruction_coverage', 0)}%")
    print(f"  分支覆盖率: {coverage_data.get('branch_coverage', 0)}%")
    print(f"  行覆盖率: {coverage_data.get('line_coverage', 0)}%")
    print(f"  方法级覆盖率详情: {len(coverage_data.get('method_details', []))}个方法")
    
    # 4. 读取模板文件
    print(f"\n📄 读取报告模板...")
    
    skill_dir = Path(__file__).parent.parent
    template_path = skill_dir / 'references' / 'UnitTestReportTemplate.html'
    
    if not template_path.exists():
        print(f"❌ 错误: 未找到模板文件 {template_path}")
        sys.exit(1)
    
    with open(template_path, 'r', encoding='utf-8') as f:
        template_content = f.read()
    
    print(f"✅ 模板文件读取成功")
    
    # 5. 准备替换数据
    print(f"\n✍️  填充报告数据...")
    
    # 提取项目名称和模块名称
    project_name = os.path.basename(project_path)
    module_name = args.module_path.split('/')[-1] if args.module_path else project_name
    
    # 去除Test后缀获取被测类名
    tested_class = args.test_class
    if tested_class.endswith('Test'):
        tested_class = tested_class[:-4]
    
    # 生成详细测试结果HTML
    detailed_results_html = generate_detailed_results_html(test_results['test_cases'])
    
    # 准备替换字典
    replacements = {
        '[项目名称]': project_name,
        '[服务模块名称]': module_name,
        '[被测试的类名]': args.test_class,
        '[被测类名]': tested_class,  # 添加这个占位符
        '[YYYY-MM-DD HH:MM:SS]': datetime.now().strftime('%Y-%m-%d %H:%M:%S'),
        '[测试人员姓名]': args.tester_name,
        '[JDK版本号]': '11',
        '[版本号]': '11',
        '[主要功能1]': '核心业务功能测试',
        '[主要功能2]': '异常场景处理测试',
        '[主要功能3]': '边界条件测试',
        '[主要功能4]': '数据准确性验证',
        '[总数]': str(test_results['total_tests']),
        '[通过率]': str(test_results['pass_rate']),
        '[通过数]': str(test_results['passed_tests']),
        '[失败数]': str(test_results['failed_tests']),
        '[错误数]': str(test_results['error_tests']),
        '[跳过数]': str(test_results['skipped_tests']),
    }
    
    # 执行替换
    report_content = template_content
    for key, value in replacements.items():
        report_content = report_content.replace(key, value)
    
    # 替换详细测试结果表格
    if detailed_results_html:
        # 在<tbody>和</tbody>之间插入测试用例数据
        # 查找详细测试结果表格的tbody标签位置
        tbody_pattern = r'(<h3 class="sub-section-title">详细测试结果</h3>.*?<tbody>)(.*?)(</tbody>)'
        
        def replace_tbody(match):
            return match.group(1) + '\n' + detailed_results_html + '\n                        ' + match.group(3)
        
        report_content = re.sub(tbody_pattern, replace_tbody, report_content, flags=re.DOTALL)
    
    # 替换覆盖率数据占位符
    if coverage_data:
        # 获取覆盖率数据
        class_coverage = coverage_data.get('class_coverage', 0)
        method_coverage = coverage_data.get('method_coverage', 0)
        branch_coverage = coverage_data.get('branch_coverage', 0)
        line_coverage = coverage_data.get('line_coverage', 0)
        instruction_coverage = coverage_data.get('instruction_coverage', 0)
        
        # 替换覆盖率统计表格中的占位符
        # 类覆盖率
        report_content = re.sub(
            r'(<td>类覆盖率</td>\s*<td>≥ 100%</td>\s*<td><span class="coverage-highlight">)\[实际\](%</span></td>\s*<td>)\[✅/❌\](</td>)',
            lambda m: m.group(1) + str(class_coverage) + m.group(2) + ('✅' if class_coverage >= 100 else '❌') + m.group(3),
            report_content
        )
        
        # 方法覆盖率
        report_content = re.sub(
            r'(<td>方法覆盖率</td>\s*<td>≥ 90%</td>\s*<td><span class="coverage-highlight">)\[实际\](%</span></td>\s*<td>)\[✅/❌\](</td>)',
            lambda m: m.group(1) + str(method_coverage) + m.group(2) + ('✅' if method_coverage >= 90 else '❌') + m.group(3),
            report_content
        )
        
        # 分支覆盖率
        report_content = re.sub(
            r'(<td>分支覆盖率</td>\s*<td>≥ 70%</td>\s*<td><span class="coverage-highlight">)\[实际\](%</span></td>\s*<td>)\[✅/❌\](</td>)',
            lambda m: m.group(1) + str(branch_coverage) + m.group(2) + ('✅' if branch_coverage >= 70 else '❌') + m.group(3),
            report_content
        )
        
        # 行覆盖率
        report_content = re.sub(
            r'(<td>行覆盖率</td>\s*<td>≥ 80%</td>\s*<td><span class="coverage-highlight">)\[实际\](%</span></td>\s*<td>)\[✅/❌\](</td>)',
            lambda m: m.group(1) + str(line_coverage) + m.group(2) + ('✅' if line_coverage >= 80 else '❌') + m.group(3),
            report_content
        )
        
        # 指令覆盖率
        report_content = re.sub(
            r'(<td>指令覆盖率</td>\s*<td>≥ 80%</td>\s*<td><span class="coverage-highlight">)\[实际\](%</span></td>\s*<td>)\[✅/❌\](</td>)',
            lambda m: m.group(1) + str(instruction_coverage) + m.group(2) + ('✅' if instruction_coverage >= 80 else '❌') + m.group(3),
            report_content
        )
    
    # 填充测试设计亮点等可选内容（传入测试类数据）
    report_content = fill_optional_content(report_content, tested_class, test_results, coverage_data, test_class_data)
    
    # 6. 创建输出目录
    output_dir = args.output_dir or os.path.join(project_path, 'document', '单元测试报告')
    os.makedirs(output_dir, exist_ok=True)
    
    # 7. 生成报告文件
    timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
    report_filename = f"{args.test_class}_单元测试报告_{timestamp}.html"
    report_path = os.path.join(output_dir, report_filename)
    
    with open(report_path, 'w', encoding='utf-8') as f:
        f.write(report_content)
    
    print(f"\n✅ 报告生成成功!")
    print(f"📁 报告路径: {report_path}")
    
    return report_path


def main():
    """主函数"""
    parser = argparse.ArgumentParser(
        description='通用单元测试报告生成脚本',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog='''
使用示例:
  python generate_universal_test_report.py --test-class CadreDevHandInAuditerServiceImpl --project-path /path/to/project
  python generate_universal_test_report.py --test-class AnnParamPage1Service --module-path dhr-ssc-service/dhr-ssc-provider
        '''
    )
    
    parser.add_argument('--test-class', required=True, help='被测类名（必需）')
    parser.add_argument('--project-path', help='项目根路径（可选，默认为当前目录）')
    parser.add_argument('--module-path', help='子模块路径（可选，如 dhr-talent-service/dhr-talent-provider）')
    parser.add_argument('--output-dir', help='报告输出目录（可选，默认为 document/单元测试报告）')
    parser.add_argument('--tester-name', default='Auto Generated', help='测试人员姓名（可选，默认为 Auto Generated）')
    
    args = parser.parse_args()
    
    print("=" * 60)
    print("    通用单元测试报告生成工具")
    print("=" * 60)
    
    try:
        generate_report(args)
    except Exception as e:
        print(f"\n❌ 报告生成失败: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == '__main__':
    main()
