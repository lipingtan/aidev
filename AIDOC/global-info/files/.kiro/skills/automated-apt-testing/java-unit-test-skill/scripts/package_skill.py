#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Java单元测试技能打包脚本
将技能打包成可分发的.skill文件
"""

import os
import zipfile
import argparse
from pathlib import Path

def validate_skill(skill_dir):
    """
    验证技能目录结构和必要文件
    """
    skill_path = Path(skill_dir)
    
    # 检查必要文件
    required_files = ['SKILL.md']
    required_dirs = ['scripts', 'references', 'assets']
    
    print("验证技能结构...")
    
    # 检查必要文件
    for file in required_files:
        file_path = skill_path / file
        if not file_path.exists():
            print(f"错误: 缺少必要文件 {file}")
            return False
        print(f"✓ 找到 {file}")
    
    # 检查必要目录
    for dir in required_dirs:
        dir_path = skill_path / dir
        if not dir_path.exists():
            print(f"错误: 缺少必要目录 {dir}")
            return False
        print(f"✓ 找到目录 {dir}")
    
    # 验证SKILL.md内容
    with open(skill_path / 'SKILL.md', 'r', encoding='utf-8') as f:
        content = f.read()
        
        if '---' not in content[:100]:
            print("错误: SKILL.md缺少YAML frontmatter")
            return False
            
        if 'name:' not in content[:500]:
            print("错误: SKILL.md缺少name字段")
            return False
            
        if 'description:' not in content[:500]:
            print("错误: SKILL.md缺少description字段")
            return False
    
    print("✓ 技能验证通过")
    return True

def create_skill_package(skill_dir, output_dir=None):
    """
    创建技能包
    """
    skill_path = Path(skill_dir)
    skill_name = skill_path.name
    
    # 确定输出路径
    if output_dir:
        output_path = Path(output_dir)
    else:
        output_path = skill_path.parent
    
    output_path.mkdir(exist_ok=True)
    
    # 创建zip文件
    skill_package = output_path / f"{skill_name}.skill"
    
    print(f"正在创建技能包: {skill_package}")
    
    with zipfile.ZipFile(skill_package, 'w', zipfile.ZIP_DEFLATED) as zipf:
        for root, dirs, files in os.walk(skill_path):
            for file in files:
                file_path = Path(root) / file
                # 计算相对路径
                rel_path = file_path.relative_to(skill_path.parent)
                zipf.write(file_path, rel_path)
    
    print(f"✓ 技能包创建完成: {skill_package}")
    return str(skill_package)

def main():
    parser = argparse.ArgumentParser(description='Java单元测试技能打包工具')
    parser.add_argument('skill_dir', help='技能目录路径')
    parser.add_argument('-o', '--output', help='输出目录路径')
    
    args = parser.parse_args()
    
    # 验证技能
    if not validate_skill(args.skill_dir):
        print("技能验证失败，无法打包")
        exit(1)
    
    # 创建技能包
    package_path = create_skill_package(args.skill_dir, args.output)
    
    print(f"\n技能包已成功创建: {package_path}")
    print("您可以将此.skill文件分享给其他人使用")

if __name__ == "__main__":
    main()
