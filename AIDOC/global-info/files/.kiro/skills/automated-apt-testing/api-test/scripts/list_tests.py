import os
import json
from pathlib import Path

TESTS_DIR = Path(__file__).parent.parent / 'tests'

def list_tests():
    if not TESTS_DIR.exists():
        print(f'测试目录不存在: {TESTS_DIR}')
        return []
    
    files = list(TESTS_DIR.iterdir())
    test_files = [f for f in files if f.suffix.lower() in ['.json', '.yaml', '.yml']]
    
    if not test_files:
        print('未找到测试文件')
        return []
    
    print('\n可用的测试文件:\n')
    for index, file in enumerate(test_files, 1):
        size = file.stat().st_size / 1024
        print(f'{index}. {file.name} ({size:.2f} KB)')
    
    return [f.name for f in test_files]

if __name__ == '__main__':
    list_tests()
