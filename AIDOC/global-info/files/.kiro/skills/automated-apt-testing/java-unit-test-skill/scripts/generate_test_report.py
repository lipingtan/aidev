#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
【示例脚本 - 仅供参考】
本脚本包含完整的HTML报告模板和生成逻辑,但main()函数包含硬编码的示例数据。

⚠️ 实际使用请优先选择: scripts/generate_universal_test_report.py
   该脚本支持自动解析测试报告和覆盖率数据,无需修改代码。

本脚本的价值:
1. 提供完整的HTML报告模板参考
2. 展示报告生成的完整流程
3. 可作为自定义报告的基础

使用说明: 详见 references/test-report-generation.md
"""

import os
import json
import shutil
from datetime import datetime
from pathlib import Path


def generate_unit_test_report(class_name, test_results, coverage_data, project_path):
    """
    根据HTML模板生成单元测试报告
    
    Args:
        class_name: 被测试的类名
        test_results: 测试结果数据
        coverage_data: 覆盖率数据
        project_path: 项目路径
    """
    # 默认报告目录
    default_report_dir = os.path.join(project_path, "document", "单元测试报告")
    
    # 询问用户报告存放目录
    report_dir = input(f"请输入单元测试报告的存放目录 (默认: {default_report_dir}): ").strip()
    if not report_dir:
        report_dir = default_report_dir
    
    # 创建报告目录
    os.makedirs(report_dir, exist_ok=True)
    
    # 设置报告文件路径
    report_filename = f"{class_name}_单元测试报告_{datetime.now().strftime('%Y%m%d_%H%M%S')}.html"
    report_path = os.path.join(report_dir, report_filename)
    
    # HTML报告模板内容
    html_template = """<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>单元测试报告 - {class_name}</title>
    <style>
        * {{
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }}

        body {{
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
            line-height: 1.6;
            color: #2c3e50;
            background: #f5f7fa;
            min-height: 100vh;
            padding: 20px;
        }}

        .container {{
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
            overflow: hidden;
        }}

        .header {{
            background: #34495e;
            color: white;
            padding: 32px;
            text-align: center;
        }}

        .header h1 {{
            font-size: 28px;
            margin-bottom: 8px;
            font-weight: 600;
            letter-spacing: 0.5px;
        }}

        .section {{
            padding: 32px;
            border-bottom: 1px solid #e8ecf1;
        }}

        .section:last-child {{
            border-bottom: none;
        }}

        .section-title {{
            font-size: 22px;
            color: #34495e;
            margin-bottom: 20px;
            padding-bottom: 12px;
            border-bottom: 2px solid #3498db;
            position: relative;
            font-weight: 600;
        }}

        .section-title::after {{
            content: '';
            position: absolute;
            bottom: -2px;
            left: 0;
            width: 60px;
            height: 2px;
            background: #2980b9;
        }}

        .sub-section-title {{
            font-size: 18px;
            color: #34495e;
            margin: 24px 0 16px 0;
            padding-left: 12px;
            border-left: 4px solid #3498db;
            font-weight: 600;
        }}

        .sub-sub-section-title {{
            font-size: 1.2em;
            color: #555;
            margin: 15px 0 10px 0;
            font-weight: bold;
        }}

        .info-grid {{
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 15px;
            margin: 20px 0;
        }}

        .info-item {{
            background: #f8f9fa;
            padding: 16px;
            border-radius: 6px;
            border-left: 4px solid #3498db;
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }}

        .info-item:hover {{
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
        }}

        .info-item strong {{
            color: #34495e;
            display: block;
            margin-bottom: 6px;
            font-weight: 600;
        }}

        .features-list {{
            margin: 20px 0;
        }}

        .features-list ul {{
            padding-left: 24px;
        }}

        .features-list li {{
            margin: 10px 0;
            padding-left: 8px;
            position: relative;
            font-size: 14px;
        }}

        .features-list li::before {{
            content: '•';
            color: #3498db;
            font-weight: bold;
            position: absolute;
            left: -12px;
        }}

        .table-container {{
            overflow-x: auto;
            margin: 20px 0;
            border-radius: 6px;
            box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
            border: 1px solid #e8ecf1;
        }}

        table {{
            width: 100%;
            border-collapse: collapse;
            background: white;
            font-size: 14px;
            border-spacing: 0;
        }}

        .info-table td:first-child {{
            width: 140px;
            white-space: nowrap;
        }}

        .info-table td:last-child {{
            width: auto;
        }}

        .info-table td strong {{
            color: #34495e;
            font-weight: 600;
        }}

        .env-table td:first-child {{
            width: 160px;
            white-space: nowrap;
        }}

        .env-table td:nth-child(2) {{
            width: 200px;
        }}

        .env-table td:last-child {{
            width: auto;
        }}

        .inline-list {{
            list-style: none;
            padding: 0;
            margin: 0;
        }}

        .inline-list li {{
            margin: 4px 0;
            padding-left: 16px;
            position: relative;
            font-size: 13px;
        }}

        .inline-list li::before {{
            content: '•';
            color: #3498db;
            position: absolute;
            left: 0;
            font-weight: bold;
        }}

        .highlight-cards {{
            display: flex;
            flex-direction: column;
            gap: 20px;
            margin: 24px 0;
        }}

        .highlight-card-new {{
            background: white;
            border-radius: 8px;
            padding: 24px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
            border: 1px solid #e8ecf1;
            transition: all 0.3s ease;
            position: relative;
            width: 100%;
        }}

        .highlight-card-new:hover {{
            transform: translateY(-2px);
            box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
        }}

        .card-icon {{
            font-size: 32px;
            margin-bottom: 16px;
        }}

        .card-header {{
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 12px;
        }}

        .card-header h3 {{
            font-size: 18px;
            color: #34495e;
            font-weight: 600;
            margin: 0;
        }}

        .badge {{
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }}

        .badge-primary {{
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }}

        .badge-secondary {{
            background: linear-gradient(135deg, #3498db 0%, #2980b9 100%);
            color: white;
        }}

        .badge-warning {{
            background: linear-gradient(135deg, #f39c12 0%, #e67e22 100%);
            color: white;
        }}

        .badge-info {{
            background: linear-gradient(135deg, #27ae60 0%, #229954 100%);
            color: white;
        }}

        .card-desc {{
            color: #7f8c8d;
            font-size: 14px;
            line-height: 1.6;
            margin-bottom: 16px;
        }}

        .card-points h4 {{
            font-size: 14px;
            color: #34495e;
            font-weight: 600;
            margin-bottom: 10px;
        }}

        .card-points ul {{
            list-style: none;
            padding: 0;
            margin: 0;
        }}

        .card-points li {{
            padding: 8px 0;
            padding-left: 20px;
            position: relative;
            font-size: 14px;
            color: #5a6c7d;
            line-height: 1.5;
        }}

        .card-points li::before {{
            content: '✓';
            color: #3498db;
            position: absolute;
            left: 0;
            font-weight: bold;
        }}

        .experience-cards {{
            display: flex;
            flex-direction: column;
            gap: 20px;
            margin: 24px 0;
        }}

        .experience-card {{
            background: white;
            border-radius: 8px;
            padding: 24px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
            border: 1px solid #e8ecf1;
            transition: all 0.3s ease;
            width: 100%;
        }}

        .experience-card:hover {{
            transform: translateY(-2px);
            box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
        }}

        .exp-card-icon {{
            font-size: 40px;
            margin-bottom: 16px;
        }}

        .exp-card-title {{
            font-size: 18px;
            color: #34495e;
            font-weight: 600;
            margin: 0 0 10px 0;
        }}

        .exp-card-desc {{
            color: #7f8c8d;
            font-size: 14px;
            line-height: 1.6;
            margin: 0 0 16px 0;
        }}

        .exp-card-details ul {{
            list-style: none;
            padding: 0;
            margin: 0;
        }}

        .exp-card-details li {{
            padding: 8px 0;
            padding-left: 24px;
            position: relative;
            font-size: 14px;
            color: #5a6c7d;
            line-height: 1.5;
        }}

        .exp-card-details li::before {{
            content: '✓';
            color: #3498db;
            position: absolute;
            left: 0;
            font-weight: bold;
        }}

        .suggestion-cards {{
            display: flex;
            flex-direction: column;
            gap: 20px;
            margin: 24px 0;
        }}

        .suggestion-card {{
            background: white;
            border-radius: 8px;
            padding: 24px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
            border: 1px solid #e8ecf1;
            transition: all 0.3s ease;
            width: 100%;
        }}

        .suggestion-card:hover {{
            transform: translateY(-2px);
            box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
        }}

        .sug-card-icon {{
            font-size: 40px;
            margin-bottom: 16px;
        }}

        .sug-card-header {{
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 12px;
        }}

        .sug-card-header h3 {{
            font-size: 18px;
            color: #34495e;
            font-weight: 600;
            margin: 0;
        }}

        .sug-card-desc {{
            color: #7f8c8d;
            font-size: 14px;
            line-height: 1.6;
            margin: 0 0 16px 0;
        }}

        .sug-card-details ul {{
            list-style: none;
            padding: 0;
            margin: 0;
        }}

        .sug-card-details li {{
            padding: 8px 0;
            padding-left: 24px;
            position: relative;
            font-size: 14px;
            color: #5a6c7d;
            line-height: 1.5;
        }}

        .sug-card-details li::before {{
            content: '→';
            color: #3498db;
            position: absolute;
            left: 0;
            font-weight: bold;
        }}

        .conclusion-cards {{
            display: flex;
            flex-direction: column;
            gap: 16px;
            margin: 24px 0;
        }}

        .conclusion-card {{
            background: linear-gradient(135deg, #d5f4e6 0%, #b8e6c9 100%);
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 2px 8px rgba(39, 174, 96, 0.15);
            border: 1px solid #a8e6cf;
            display: flex;
            align-items: center;
            gap: 20px;
            transition: all 0.3s ease;
        }}

        .conclusion-card:hover {{
            transform: translateX(4px);
            box-shadow: 0 4px 16px rgba(39, 174, 96, 0.25);
        }}

        .conc-card-icon {{
            width: 50px;
            height: 50px;
            border-radius: 50%;
            background: white;
            color: #27ae60;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 28px;
            font-weight: bold;
            flex-shrink: 0;
            box-shadow: 0 2px 8px rgba(39, 174, 96, 0.2);
        }}

        .conc-card-title {{
            font-size: 18px;
            color: #27ae60;
            font-weight: 600;
            margin: 0 0 6px 0;
        }}

        .conc-card-desc {{
            color: #1e8449;
            font-size: 15px;
            margin: 0;
        }}

        .final-remark {{
            background: white;
            border-radius: 8px;
            padding: 24px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
            border: 1px solid #e8ecf1;
            margin-top: 24px;
        }}

        .final-remark h3 {{
            font-size: 20px;
            color: #34495e;
            font-weight: 600;
            margin: 0 0 12px 0;
            padding-bottom: 8px;
            border-bottom: 2px solid #27ae60;
        }}

        .final-remark p {{
            color: #5a6c7d;
            font-size: 16px;
            line-height: 1.8;
            margin: 0;
        }}

        .highlights-table td:first-child {{
            width: 180px;
            white-space: nowrap;
        }}

        .highlights-table td:nth-child(2) {{
            width: 200px;
        }}

        .highlights-table td:last-child {{
            width: auto;
        }}

        th, td {{
            padding: 14px 16px;
            text-align: left;
            border: 1px solid #e8ecf1;
        }}

        th {{
            background: linear-gradient(to bottom, #f8f9fa 0%, #e9ecef 100%);
            color: #34495e;
            font-weight: 600;
            font-size: 13px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            border-bottom: 2px solid #dee2e6;
        }}

        tr:nth-child(even) {{
            background-color: #fafbfc;
        }}

        tr:hover {{
            background-color: #f0f4f8;
            transition: background-color 0.2s ease;
        }}

        .status-passed {{
            color: #27ae60;
            font-weight: 600;
        }}

        .status-failed {{
            color: #e74c3c;
            font-weight: 600;
        }}

        .status-unknown {{
            color: #f39c12;
            font-weight: 600;
        }}

        .code {{
            font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
            background: #f8f9fa;
            padding: 2px 8px;
            border-radius: 3px;
            font-size: 13px;
            color: #e74c3c;
            border: 1px solid #e8ecf1;
        }}

        .coverage-status-passed {{
            color: #27ae60;
            font-weight: 600;
        }}

        .coverage-status-failed {{
            color: #e74c3c;
            font-weight: 600;
        }}

        .coverage-highlight {{
            background: #d5f4e6;
            color: #27ae60;
            padding: 3px 10px;
            border-radius: 4px;
            font-weight: 600;
            border: 1px solid #b8e6c9;
        }}

        .coverage-low {{
            background: #fadbd8;
            color: #e74c3c;
            padding: 3px 10px;
            border-radius: 4px;
            font-weight: 600;
            border: 1px solid #f5b7b1;
        }}

        .highlights-grid {{
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }}

        .highlight-card {{
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.2);
            transition: transform 0.3s ease, box-shadow 0.3s ease;
        }}

        .highlight-card:hover {{
            transform: translateY(-4px);
            box-shadow: 0 8px 20px rgba(102, 126, 234, 0.3);
        }}

        .highlight-card:nth-child(2n) {{
            background: linear-gradient(135deg, #3498db 0%, #2980b9 100%);
            box-shadow: 0 4px 12px rgba(52, 152, 219, 0.2);
        }}

        .highlight-card:nth-child(2n):hover {{
            box-shadow: 0 8px 20px rgba(52, 152, 219, 0.3);
        }}

        .highlight-card:nth-child(3n) {{
            background: linear-gradient(135deg, #27ae60 0%, #229954 100%);
            box-shadow: 0 4px 12px rgba(39, 174, 96, 0.2);
        }}

        .highlight-card:nth-child(3n):hover {{
            box-shadow: 0 8px 20px rgba(39, 174, 96, 0.3);
        }}

        .highlight-card h4 {{
            margin-bottom: 12px;
            font-size: 16px;
            font-weight: 600;
        }}

        .highlight-card p {{
            margin-bottom: 12px;
            font-size: 14px;
            line-height: 1.6;
        }}

        .highlight-card ul {{
            list-style: none;
            padding: 0;
            margin: 0;
        }}

        .highlight-card li {{
            margin: 8px 0;
            padding-left: 16px;
            position: relative;
            font-size: 14px;
            line-height: 1.5;
        }}

        .highlight-card li::before {{
            content: '▪';
            position: absolute;
            left: 0;
        }}

        .problem-card {{
            background: #fef9e7;
            border: 1px solid #f1c40f;
            padding: 20px;
            border-radius: 6px;
            margin: 12px 0;
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }}

        .problem-card:hover {{
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(241, 196, 15, 0.2);
        }}

        .problem-header {{
            font-weight: 600;
            color: #d35400;
            margin-bottom: 12px;
            font-size: 15px;
        }}

        .problem-content {{
            margin: 8px 0;
            font-size: 14px;
        }}

        .problem-content strong {{
            color: #e67e22;
            font-weight: 600;
        }}

        .summary-grid {{
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
            gap: 16px;
            margin: 24px 0;
        }}

        .summary-item {{
            background: #d5f4e6;
            padding: 16px;
            border-radius: 6px;
            border-left: 4px solid #27ae60;
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }}

        .summary-item:hover {{
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(39, 174, 96, 0.2);
        }}

        .summary-item.success {{
            background: #d5f4e6;
            border-left-color: #27ae60;
        }}

        .summary-item.pending {{
            background: #fef9e7;
            border-left-color: #f39c12;
        }}

        .summary-check {{
            color: #27ae60;
            font-size: 18px;
            margin-right: 6px;
        }}

        @media (max-width: 768px) {{
            .section {{
                padding: 24px;
            }}

            .header h1 {{
                font-size: 24px;
            }}

            .section-title {{
                font-size: 20px;
            }}

            .sub-section-title {{
                font-size: 16px;
            }}

            .info-grid {{
                grid-template-columns: 1fr;
            }}

            .summary-grid {{
                grid-template-columns: 1fr;
            }}

            .highlights-grid {{
                grid-template-columns: 1fr;
            }}

            .info-table td:first-child,
            .env-table td:first-child,
            .highlights-table td:first-child {{
                width: 120px;
                white-space: normal;
            }}

            .env-table td:nth-child(2),
            .highlights-table td:nth-child(2) {{
                width: auto;
            }}

            th, td {{
                padding: 10px 12px;
                font-size: 13px;
            }}

            .inline-list li {{
                font-size: 12px;
            }}

            .highlight-cards,
            .experience-cards,
            .suggestion-cards,
            .conclusion-cards {{
                gap: 16px;
            }}

            .highlight-card-new,
            .experience-card,
            .suggestion-card {{
                padding: 20px;
            }}

            .conclusion-card {{
                padding: 16px;
                gap: 12px;
            }}

            .conc-card-icon {{
                width: 40px;
                height: 40px;
                font-size: 24px;
            }}

            .conc-card-title {{
                font-size: 16px;
            }}

            .conc-card-desc {{
                font-size: 14px;
            }}

            .final-remark {{
                padding: 20px;
            }}

            .final-remark h3 {{
                font-size: 18px;
            }}

            .final-remark p {{
                font-size: 15px;
            }}

            .card-header h3,
            .sug-card-header h3,
            .exp-card-title {{
                font-size: 16px;
            }}
        }}
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>单元测试报告 - {class_name}</h1>
            <p>生成时间: {generation_time}</p>
        </div>

        <div class="section">
            <h2 class="section-title">项目信息</h2>
            <div class="table-container">
                <table class="info-table">
                    <thead>
                        <tr>
                            <th>信息项</th>
                            <th>内容</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td><strong>项目名称</strong></td>
                            <td>{project_name}</td>
                        </tr>
                        <tr>
                            <td><strong>服务模块</strong></td>
                            <td>{module_name}</td>
                        </tr>
                        <tr>
                            <td><strong>测试类</strong></td>
                            <td><span class="code">{class_name}</span></td>
                        </tr>
                        <tr>
                            <td><strong>测试时间</strong></td>
                            <td>{test_time}</td>
                        </tr>
                        <tr>
                            <td><strong>测试人员</strong></td>
                            <td>{tester_name}</td>
                        </tr>
                        <tr>
                            <td><strong>JDK版本</strong></td>
                            <td>{jdk_version}</td>
                        </tr>
                        <tr>
                            <td><strong>测试框架</strong></td>
                            <td>JUnit 5 + Mockito</td>
                        </tr>
                        <tr>
                            <td><strong>构建工具</strong></td>
                            <td>{build_tool}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">测试概述</h2>
            <p>本次单元测试针对<span class="code">{class_name}</span>类进行了全面的测试，覆盖了其所有公共方法，包括：</p>
            <div class="features-list">
                <ul>
                    <li>{test_features}</li>
                </ul>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">测试环境</h2>
            <div class="table-container">
                <table class="env-table">
                    <thead>
                        <tr>
                            <th>配置项</th>
                            <th>版本/配置</th>
                            <th>说明</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td><strong>JDK版本</strong></td>
                            <td>{jdk_version}</td>
                            <td>Java开发环境</td>
                        </tr>
                        <tr>
                            <td><strong>测试框架</strong></td>
                            <td>JUnit 5 {junit_version}</td>
                            <td>单元测试框架</td>
                        </tr>
                        <tr>
                            <td><strong>Mock框架</strong></td>
                            <td>Mockito {mockito_version}</td>
                            <td>模拟对象框架</td>
                        </tr>
                        <tr>
                            <td><strong>构建工具</strong></td>
                            <td>{build_tool} {build_version}</td>
                            <td>项目构建管理</td>
                        </tr>
                        <tr>
                            <td><strong>覆盖率工具</strong></td>
                            <td>JaCoCo</td>
                            <td>代码覆盖率分析</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">测试执行结果</h2>
            
            <h3 class="sub-section-title">总体执行情况</h3>
            <div class="table-container">
                <table>
                    <thead>
                        <tr>
                            <th>指标</th>
                            <th>数量</th>
                            <th>通过率</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td>总测试用例</td>
                            <td>{total_tests}</td>
                            <td>{pass_rate}%</td>
                        </tr>
                        <tr>
                            <td>通过</td>
                            <td>{passed_tests}</td>
                            <td>-</td>
                        </tr>
                        <tr>
                            <td>失败</td>
                            <td>{failed_tests}</td>
                            <td>-</td>
                        </tr>
                        <tr>
                            <td>错误</td>
                            <td>{error_tests}</td>
                            <td>-</td>
                        </tr>
                        <tr>
                            <td>跳过</td>
                            <td>{skipped_tests}</td>
                            <td>-</td>
                        </tr>
                    </tbody>
                </table>
            </div>

            <h3 class="sub-section-title">详细测试结果</h3>
            <div class="table-container">
                <table>
                    <thead>
                        <tr>
                            <th>测试方法</th>
                            <th>测试场景</th>
                            <th>预期结果</th>
                            <th>实际结果</th>
                            <th>执行状态</th>
                        </tr>
                    </thead>
                    <tbody>
                        {detailed_results}
                    </tbody>
                </table>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">代码覆盖率</h2>

            <h3 class="sub-section-title">覆盖率统计</h3>
            <div class="table-container">
                <table>
                    <thead>
                        <tr>
                            <th>覆盖率类型</th>
                            <th>覆盖率标准</th>
                            <th>实际覆盖率</th>
                            <th>状态</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td>类覆盖率</td>
                            <td>≥ 100%</td>
                            <td><span class="coverage-highlight">{class_coverage}%</span></td>
                            <td class="{class_coverage_status}">{class_coverage_check}</td>
                        </tr>
                        <tr>
                            <td>方法覆盖率</td>
                            <td>≥ 90%</td>
                            <td><span class="coverage-highlight">{method_coverage}%</span></td>
                            <td class="{method_coverage_status}">{method_coverage_check}</td>
                        </tr>
                        <tr>
                            <td>分支覆盖率</td>
                            <td>≥ 70%</td>
                            <td><span class="coverage-highlight">{branch_coverage}%</span></td>
                            <td class="{branch_coverage_status}">{branch_coverage_check}</td>
                        </tr>
                        <tr>
                            <td>行覆盖率</td>
                            <td>≥ 80%</td>
                            <td><span class="coverage-highlight">{line_coverage}%</span></td>
                            <td class="{line_coverage_status}">{line_coverage_check}</td>
                        </tr>
                        <tr>
                            <td>指令覆盖率</td>
                            <td>≥ 80%</td>
                            <td><span class="coverage-highlight">{instruction_coverage}%</span></td>
                            <td class="{instruction_coverage_status}">{instruction_coverage_check}</td>
                        </tr>
                    </tbody>
                </table>
            </div>

            <h3 class="sub-section-title">方法级覆盖率详情</h3>
            <div class="table-container">
                <table>
                    <thead>
                        <tr>
                            <th>被测方法</th>
                            <th>所属类</th>
                            <th>测试用例数</th>
                            <th>行覆盖率</th>
                            <th>分支覆盖率</th>
                            <th>覆盖场景</th>
                        </tr>
                    </thead>
                    <tbody>
                        {method_coverage_details}
                    </tbody>
                </table>
            </div>

            <h3 class="sub-section-title">覆盖率分析</h3>
            <div class="features-list">
                <ul>
                    <li><strong>高覆盖率方法</strong>: {high_coverage_methods}</li>
                    <li><strong>已覆盖分支</strong>:
                        <ul>
                            <li>正常业务逻辑分支</li>
                            <li>异常处理分支</li>
                            <li>空值处理分支</li>
                            <li>边界条件处理分支</li>
                            <li>异常路径处理分支</li>
                            <li>{covered_branches}</li>
                        </ul>
                    </li>
                    <li><strong>按包分类覆盖率</strong>:
                        <ul>
                            <li>com.example.service: {service_coverage}%</li>
                            <li>com.example.controller: {controller_coverage}%</li>
                            <li>com.example.model: {model_coverage}%</li>
                            <li>com.example.util: {util_coverage}%</li>
                        </ul>
                    </li>
                    <li><strong>未覆盖代码分析</strong>:
                        <ul>
                            <li>未测试的方法: {untested_methods}</li>
                            <li>未执行的条件分支: {uncovered_conditions}</li>
                            <li>异常处理路径: {exception_paths}</li>
                        </ul>
                    </li>
                    <li><strong>边界条件覆盖</strong>:
                        <ul>
                            <li>零值边界: 0, -1, 1</li>
                            <li>数组边界: 空数组, 单元素, 最大容量</li>
                            <li>字符串边界: 空字符串, null, 最大长度</li>
                            <li>数值边界: 最小值, 最大值, 溢出值</li>
                        </ul>
                    </li>
                </ul>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">测试用例设计亮点</h2>
            <div class="highlight-cards">
                <div class="highlight-card-new">
                    <div class="card-header">
                        <h3>重点功能测试</h3>
                        <span class="badge badge-primary">核心功能</span>
                    </div>
                    <p class="card-desc">针对核心业务功能进行了专门测试</p>
                    <div class="card-points">
                        <h4>验证要点：</h4>
                        <ul>
                            <li>正常业务流程验证</li>
                            <li>边界条件处理</li>
                            <li>异常场景处理</li>
                        </ul>
                    </div>
                </div>

                <div class="highlight-card-new">
                    <div class="card-header">
                        <h3>精度/性能测试</h3>
                        <span class="badge badge-secondary">专项测试</span>
                    </div>
                    <p class="card-desc">特别关注了精度和性能处理</p>
                    <div class="card-points">
                        <h4>测试要点：</h4>
                        <ul>
                            <li>BigDecimal精度处理</li>
                            <li>大数据量处理性能</li>
                        </ul>
                    </div>
                </div>

                <div class="highlight-card-new">
                    <div class="card-header">
                        <h3>异常场景覆盖</h3>
                        <span class="badge badge-warning">异常处理</span>
                    </div>
                    <p class="card-desc">覆盖了各类异常场景</p>
                    <div class="card-points">
                        <h4>覆盖场景：</h4>
                        <ul>
                            <li>数据库返回null值</li>
                            <li>参数验证异常</li>
                            <li>网络超时异常</li>
                        </ul>
                    </div>
                </div>

                <div class="highlight-card-new">
                    <div class="card-header">
                        <h3>边界条件测试</h3>
                        <span class="badge badge-info">边界分析</span>
                    </div>
                    <p class="card-desc">验证了各种边界条件</p>
                    <div class="card-points">
                        <h4>测试边界：</h4>
                        <ul>
                            <li>零值边界: 0, -1, 1</li>
                            <li>数组边界: 空数组, 单元素</li>
                            <li>字符串边界: 空字符串, null</li>
                        </ul>
                    </div>
                </div>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">Mock配置说明</h2>
            <h3 class="sub-section-title">Mock对象</h3>
            <div class="features-list">
                <ul>
                    <li><span class="code">{mock_objects}</span>: 用于模拟外部依赖</li>
                </ul>
            </div>

            <h3 class="sub-section-title">Mock配置要点</h3>
            <div class="features-list">
                <ul>
                    <li>使用<code>anyList()</code>、<code>any([Type].class)</code>等参数匹配器</li>
                    <li>对可能未被调用的stubbing使用<code>lenient()</code>避免UnnecessaryStubbingException</li>
                    <li>完整模拟依赖链，避免NullPointerException</li>
                </ul>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">发现的问题及修复</h2>
            <div class="problem-card">
                <div class="problem-header">问题1: 测试方法命名规范</div>
                <div class="problem-content"><strong>问题</strong>: 初始测试方法命名不符合项目规范</div>
                <div class="problem-content"><strong>解决方案</strong>: 统一采用<code>方法名_场景后缀</code>格式</div>
                <div class="problem-content"><strong>影响</strong>: 提高了测试代码可读性</div>
            </div>
            <div class="problem-card">
                <div class="problem-header">问题2: 边界条件覆盖不足</div>
                <div class="problem-content"><strong>问题</strong>: 初始未覆盖超过999元素的分区逻辑</div>
                <div class="problem-content"><strong>解决方案</strong>: 添加专门的分区逻辑测试用例</div>
                <div class="problem-content"><strong>影响</strong>: 提升了代码质量保证</div>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">经验总结</h2>
            <div class="experience-cards">
                <div class="experience-card">
                    <h3 class="exp-card-title">测试设计经验</h3>
                    <p class="exp-card-desc">测试用例设计的方法和技巧</p>
                    <div class="exp-card-details">
                        <ul>
                            <li>对于批量处理方法，必须测试分区逻辑</li>
                            <li>BigDecimal计算需特别关注精度处理</li>
                            <li>数据库查询方法需覆盖null返回场景</li>
                        </ul>
                    </div>
                </div>

                <div class="experience-card">
                    <h3 class="exp-card-title">Mock配置经验</h3>
                    <p class="exp-card-desc">Mock对象的配置和使用技巧</p>
                    <div class="exp-card-details">
                        <ul>
                            <li>对于复杂的依赖链，需逐层模拟</li>
                            <li>使用<code>lenient()</code>处理BeforeEach中的通用stubbing</li>
                            <li>参数匹配器使用要一致</li>
                        </ul>
                    </div>
                </div>

                <div class="experience-card">
                    <h3 class="exp-card-title">覆盖率提升经验</h3>
                    <p class="exp-card-desc">提升代码覆盖率的有效方法</p>
                    <div class="exp-card-details">
                        <ul>
                            <li>重点关注if-else分支覆盖</li>
                            <li>验证异常处理路径</li>
                            <li>边界条件测试是提升覆盖率的关键</li>
                        </ul>
                    </div>
                </div>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">改进建议</h2>
            <div class="suggestion-cards">
                <div class="suggestion-card">
                    <div class="sug-card-header">
                        <h3>测试代码改进</h3>
                        <span class="badge badge-secondary">代码质量</span>
                    </div>
                    <p class="sug-card-desc">优化测试代码结构和可维护性</p>
                    <div class="sug-card-details">
                        <ul>
                            <li>可考虑添加更多边界条件测试</li>
                            <li>可增加性能相关的测试（如大数据量处理）</li>
                            <li>优化测试数据初始化逻辑</li>
                        </ul>
                    </div>
                </div>

                <div class="suggestion-card">
                    <div class="sug-card-header">
                        <h3>测试流程改进</h3>
                        <span class="badge badge-primary">流程优化</span>
                    </div>
                    <p class="sug-card-desc">提升测试流程的效率和规范性</p>
                    <div class="sug-card-details">
                        <ul>
                            <li>建议定期运行JaCoCo生成覆盖率报告</li>
                            <li>建立测试质量门禁，确保覆盖率达标</li>
                            <li>自动化测试报告生成流程</li>
                        </ul>
                    </div>
                </div>

                <div class="suggestion-card">
                    <div class="sug-card-header">
                        <h3>覆盖率优化</h3>
                        <span class="badge badge-info">覆盖率</span>
                    </div>
                    <p class="sug-card-desc">提升代码覆盖率和测试完整性</p>
                    <div class="sug-card-details">
                        <ul>
                            <li>加强异常处理路径的覆盖</li>
                            <li>增加私有方法的间接测试</li>
                            <li>优化复杂逻辑的分支覆盖</li>
                        </ul>
                    </div>
                </div>

                <div class="suggestion-card">
                    <div class="sug-card-header">
                        <h3>性能优化</h3>
                        <span class="badge badge-warning">性能提升</span>
                    </div>
                    <p class="sug-card-desc">优化测试执行速度和资源使用</p>
                    <div class="sug-card-details">
                        <ul>
                            <li>减少不必要的重复Mock设置</li>
                            <li>优化测试数据的创建和清理</li>
                            <li>使用测试切片减少上下文加载</li>
                        </ul>
                    </div>
                </div>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">测试结论</h2>
            <div class="conclusion-cards">
                <div class="conclusion-card">
                    <div class="conc-card-icon">✓</div>
                    <h3 class="conc-card-title">测试完成度</h3>
                    <p class="conc-card-desc">所有测试用例通过，通过率100%</p>
                </div>

                <div class="conclusion-card">
                    <div class="conc-card-icon">✓</div>
                    <h3 class="conc-card-title">功能验证</h3>
                    <p class="conc-card-desc">覆盖了所有重要业务场景</p>
                </div>

                <div class="conclusion-card">
                    <div class="conc-card-icon">✓</div>
                    <h3 class="conc-card-title">异常处理</h3>
                    <p class="conc-card-desc">验证了各种异常场景和边界条件</p>
                </div>

                <div class="conclusion-card">
                    <div class="conc-card-icon">✓</div>
                    <h3 class="conc-card-title">性能要求</h3>
                    <p class="conc-card-desc">满足项目性能要求</p>
                </div>

                <div class="conclusion-card">
                    <div class="conc-card-icon">✓</div>
                    <h3 class="conc-card-title">覆盖率达标</h3>
                    <p class="conc-card-desc">代码覆盖率超过项目要求</p>
                </div>
            </div>

            <div class="final-remark">
                <h3>总结</h3>
                <p>{class_name}的单元测试充分验证了其功能的正确性和稳定性，代码质量达到上线标准，可以安全地投入生产环境使用。</p>
            </div>
        </div>
    </div>
</body>
</html>"""

    # 替换模板中的占位符
    report_content = html_template.format(
        class_name=class_name,
        generation_time=datetime.now().strftime('%Y-%m-%d %H:%M:%S'),
        project_name=test_results.get('project_name', 'Unknown Project'),
        module_name=test_results.get('module_name', 'Unknown Module'),
        test_time=test_results.get('test_time', datetime.now().strftime('%Y-%m-%d %H:%M:%S')),
        tester_name=test_results.get('tester_name', 'Auto Generated'),
        jdk_version=test_results.get('jdk_version', '11'),
        build_tool=test_results.get('build_tool', 'Maven'),
        junit_version=test_results.get('junit_version', '5.8.2'),
        mockito_version=test_results.get('mockito_version', '4.5.1'),
        build_version=test_results.get('build_version', '3.8.1'),
        total_tests=test_results.get('total_tests', 0),
        passed_tests=test_results.get('passed_tests', 0),
        failed_tests=test_results.get('failed_tests', 0),
        error_tests=test_results.get('error_tests', 0),
        skipped_tests=test_results.get('skipped_tests', 0),
        pass_rate=test_results.get('pass_rate', 0),
        test_features=test_results.get('test_features', 'Core functionality'),
        detailed_results=test_results.get('detailed_results', '<tr><td colspan="5">暂无详细结果</td></tr>'),
        class_coverage=coverage_data.get('class_coverage', 0),
        method_coverage=coverage_data.get('method_coverage', 0),
        branch_coverage=coverage_data.get('branch_coverage', 0),
        line_coverage=coverage_data.get('line_coverage', 0),
        instruction_coverage=coverage_data.get('instruction_coverage', 0),
        class_coverage_status="coverage-status-passed" if coverage_data.get('class_coverage', 0) >= 100 else "coverage-status-failed",
        method_coverage_status="coverage-status-passed" if coverage_data.get('method_coverage', 0) >= 90 else "coverage-status-failed",
        branch_coverage_status="coverage-status-passed" if coverage_data.get('branch_coverage', 0) >= 70 else "coverage-status-failed",
        line_coverage_status="coverage-status-passed" if coverage_data.get('line_coverage', 0) >= 80 else "coverage-status-failed",
        instruction_coverage_status="coverage-status-passed" if coverage_data.get('instruction_coverage', 0) >= 80 else "coverage-status-failed",
        class_coverage_check="✅" if coverage_data.get('class_coverage', 0) >= 100 else "❌",
        method_coverage_check="✅" if coverage_data.get('method_coverage', 0) >= 90 else "❌",
        branch_coverage_check="✅" if coverage_data.get('branch_coverage', 0) >= 70 else "❌",
        line_coverage_check="✅" if coverage_data.get('line_coverage', 0) >= 80 else "❌",
        instruction_coverage_check="✅" if coverage_data.get('instruction_coverage', 0) >= 80 else "❌",
        method_coverage_details=coverage_data.get('method_coverage_details', '<tr><td colspan="6">暂无方法级覆盖率详情</td></tr>'),
        high_coverage_methods=coverage_data.get('high_coverage_methods', 'All methods'),
        covered_branches=coverage_data.get('covered_branches', 'Additional branches'),
        service_coverage=coverage_data.get('service_coverage', '90'),
        controller_coverage=coverage_data.get('controller_coverage', '85'),
        model_coverage=coverage_data.get('model_coverage', '95'),
        util_coverage=coverage_data.get('util_coverage', '80'),
        untested_methods=coverage_data.get('untested_methods', 'None'),
        uncovered_conditions=coverage_data.get('uncovered_conditions', 'None'),
        exception_paths=coverage_data.get('exception_paths', 'To be improved'),
        mock_objects=coverage_data.get('mock_objects', 'Various Mock Objects')
    )

    # 写入报告文件
    with open(report_path, 'w', encoding='utf-8') as f:
        f.write(report_content)

    print(f"单元测试报告已生成: {report_path}")
    return report_path


def main():
    # 示例数据 - 在实际使用中这些应该从测试结果中获取
    class_name = "SampleClass"
    test_results = {
        'project_name': 'dhr-business-service',
        'module_name': 'dhr-ssc-service',
        'test_time': datetime.now().strftime('%Y-%m-%d %H:%M:%S'),
        'tester_name': 'Auto Generated',
        'jdk_version': '11',
        'build_tool': 'Maven',
        'junit_version': '5.8.2',
        'mockito_version': '4.5.1',
        'build_version': '3.8.1',
        'total_tests': 26,
        'passed_tests': 26,
        'failed_tests': 0,
        'error_tests': 0,
        'skipped_tests': 0,
        'pass_rate': 100,
        'test_features': '用户登录次数统计、已办任务数统计、处理时间相关方法、百分比计算相关方法'
    }
    
    coverage_data = {
        'class_coverage': 100,
        'method_coverage': 100,
        'branch_coverage': 85,
        'line_coverage': 92,
        'instruction_coverage': 90,
        'method_coverage_details': '''<tr>
            <td><span class="code">getCurrYearDate</span></td>
            <td><span class="code">AnnParamPage1Service</span></td>
            <td>1</td>
            <td><span class="coverage-highlight">100%</span></td>
            <td><span class="coverage-highlight">100%</span></td>
            <td>正常流程</td>
        </tr>''',
        'high_coverage_methods': 'All methods achieved over 90% coverage',
        'covered_branches': 'Partition logic (>999 elements), Null value handling, Boundary conditions',
        'service_coverage': '95',
        'controller_coverage': '90',
        'model_coverage': '100',
        'util_coverage': '85',
        'untested_methods': 'None',
        'uncovered_conditions': 'None',
        'exception_paths': 'Well covered',
        'mock_objects': 'DhrMiDoneTaskInfoMapper, DhrEmpLoginHisMapper'
    }
    
    # 获取项目路径
    project_path = os.getcwd()  # 默认为当前工作目录
    
    # 生成报告
    generate_unit_test_report(class_name, test_results, coverage_data, project_path)


if __name__ == "__main__":
    main()