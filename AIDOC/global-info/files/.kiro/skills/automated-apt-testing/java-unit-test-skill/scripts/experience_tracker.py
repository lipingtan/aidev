#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
问题修复绋验积累追踪器

⚠️ 重要声明：
    本工具积累的是「问题修复的通用经验」，而不是「每次测试的执行结果」
    
    ✅ 应该积累：
        - 编译错误的根本原因和解决方案
        - 测试失败的Mockito配置技巧
        - 框架兼容性问题的处理方法
        - 通用的测试模式（分区逻辑、精度计算）
    
    ❌ 不应该积累：
        - 某次测试的执行结果（多少个测试通过/失败）
        - 特定项目特定类的业务逻辑细节

使用方法：
    1. 当修复一个问题后，调用 record_experience()
    2. 脚本会自动显示积累记录的文本
    3. 告诉用户问题已被总结并记录
"""

import json
import os
from datetime import datetime
from typing import List, Dict, Optional

class ExperienceTracker:
    """问题修复经验积累追踪器"""
    
    def __init__(self, skill_dir: str = '.'):
        """
        初始化追踪器
        
        Args:
            skill_dir: skill目录路径
        """
        self.skill_dir = skill_dir
        self.tracker_file = os.path.join(skill_dir, '.experience_tracker.json')
        self.guide_file = os.path.join(skill_dir, 'references/troubleshooting-guide.md')
        self.load_tracker()
    
    def load_tracker(self):
        """加载追踪状态"""
        if os.path.exists(self.tracker_file):
            with open(self.tracker_file, 'r', encoding='utf-8') as f:
                self.data = json.load(f)
        else:
            self.data = {
                'current_session': None,
                'accumulated_experiences': [],
                'last_guide_update': None
            }
    
    def save_tracker(self):
        """保存追踪状态"""
        with open(self.tracker_file, 'w', encoding='utf-8') as f:
            json.dump(self.data, f, ensure_ascii=False, indent=2)
    
    def start_session(self, test_class_name: str):
        """
        开始新的测试会话
        
        Args:
            test_class_name: 测试类名
        """
        self.data['current_session'] = {
            'test_class': test_class_name,
            'start_time': datetime.now().isoformat(),
            'experiences_count': 0
        }
        self.save_tracker()
        
        print(f"\n✅ 已启动测试会话: {test_class_name}")
        print(f"开始时间: {self.data['current_session']['start_time']}")
        print("-" * 60 + "\n")
    
    def record_experience(self, 
                         problem_type: str, 
                         problem_summary: str,
                         root_cause: str,
                         solution: str,
                         target_section: str) -> str:
        """
        记录问题修复经验（自动显示积累文本）
        
        Args:
            problem_type: 问题类型 ('compile_error' | 'test_failure' | 'framework_issue' | 'mock_config')
            problem_summary: 问题摘要
            root_cause: 根本原因
            solution: 解决方案
            target_section: 目标章节（常见问题速查表/Mockito异常处理/测试模式）
            
        Returns:
            str: 积累记录的文本
        """
        experience = {
            'type': problem_type,
            'summary': problem_summary,
            'root_cause': root_cause,
            'solution': solution,
            'target_section': target_section,
            'recorded_time': datetime.now().isoformat(),
            'test_class': self.data['current_session']['test_class'] if self.data['current_session'] else 'Unknown'
        }
        
        self.data['accumulated_experiences'].append(experience)
        if self.data['current_session']:
            self.data['current_session']['experiences_count'] += 1
        self.data['last_guide_update'] = datetime.now().isoformat()
        self.save_tracker()
        
        # 自动显示积累记录的文本
        return self._display_experience_record(experience)
    
    def _display_experience_record(self, experience: Dict) -> str:
        """
        显示积累记录的文本
        
        Args:
            experience: 经验记录
            
        Returns:
            str: 格式化的显示文本
        """
        type_names = {
            'compile_error': '编译错误',
            'test_failure': '测试失败',
            'framework_issue': '框架问题',
            'mock_config': 'Mock配置'
        }
        
        display_text = f"""
{"="*80}
📝 问题修复经验已积累
{"="*80}

📌 问题类型：{type_names.get(experience['type'], experience['type'])}
📝 问题摘要：{experience['summary']}
🔍 根本原因：{experience['root_cause']}
✅ 解决方案：{experience['solution']}

📚 记录位置：troubleshooting-guide.md → 「{experience['target_section']}」
⏰ 记录时间：{experience['recorded_time']}

🎯 价值说明：
   - 此经验适用于所有使用 JUnit 5 + Mockito 的 Java 项目
   - 下次遇到类似问题时，可以直接应用此解决方案
   - 在生成测试代码前，查阅此经验可以预防重复犯错

{"="*80}
"""
        print(display_text)
        return display_text
    
    def get_session_summary(self) -> str:
        """
        获取当前会话的积累总结
        
        Returns:
            str: 格式化的总结文本
        """
        if not self.data['current_session']:
            return "❌ 无活跃会话"
        
        session = self.data['current_session']
        
        # 获取本次会话积累的经验
        session_experiences = [
            exp for exp in self.data['accumulated_experiences']
            if exp.get('test_class') == session['test_class']
        ]
        
        summary = f"""
{"="*80}
📊 本次测试会话积累总结
{"="*80}

📌 测试类: {session['test_class']}
⏰ 开始时间: {session['start_time']}
⏱️  当前时间: {datetime.now().isoformat()}

📚 积累经验数量: {len(session_experiences)} 个

"""
        
        if session_experiences:
            summary += "✅ 已积累的经验：\n"
            for i, exp in enumerate(session_experiences, 1):
                summary += f"   {i}. [{exp['type']}] {exp['summary']}\n"
                summary += f"      → 记录位置: {exp['target_section']}\n"
            summary += "\n"
        else:
            summary += "📌 本次测试未遇到新问题，无需积累新经验\n\n"
        
        summary += f"""
🎯 积累价值：
   - 本次积累的经验适用于所有 Java 单元测试项目
   - 已记录到 troubleshooting-guide.md，下次可以直接应用
   - 形成“预防→修复→积累→预防”的持续改进闭环

{"="*80}
"""
        
        return summary
    
    def end_session(self):
        """结束当前会话"""
        if self.data['current_session']:
            summary = self.get_session_summary()
            print(summary)
            self.data['current_session'] = None
            self.save_tracker()
    
    def get_all_experiences(self) -> List[Dict]:
        """
        获取所有积累的经验
        
        Returns:
            List[Dict]: 经验列表
        """
        return self.data['accumulated_experiences']
    
    def search_experience(self, keyword: str) -> List[Dict]:
        """
        搜索相关经验
        
        Args:
            keyword: 搜索关键词
            
        Returns:
            List[Dict]: 匹配的经验列表
        """
        results = []
        for exp in self.data['accumulated_experiences']:
            if (keyword.lower() in exp['summary'].lower() or 
                keyword.lower() in exp['root_cause'].lower() or 
                keyword.lower() in exp['solution'].lower()):
                results.append(exp)
        return results


# 使用示例
if __name__ == '__main__':
    # 初始化追踪器
    tracker = ExperienceTracker('.')
    
    # 开始测试会话
    tracker.start_session('CadreDevHandInAuditerServiceImplTest')
    
    # 示例1：记录编译错误修复经验
    print("\n\n=== 示例1：记录编译错误修复经验 ===")
    tracker.record_experience(
        problem_type='compile_error',
        problem_summary='静态Mock不支持',
        root_cause='缺少 mockito-inline 依赖，无法Mock静态方法',
        solution='在 pom.xml 中添加 mockito-inline 依赖',
        target_section='常见问题速查表'
    )
    
    # 示例2：记录测试失败修复经验
    print("\n\n=== 示例2：记录测试失败修复经验 ===")
    tracker.record_experience(
        problem_type='test_failure',
        problem_summary='ResponseVO导致StackOverflow',
        root_cause='ResponseVO.success()内部调用MessageUtils依赖Spring容器，导致无限递归',
        solution='使用 doReturn().when() 模式，先mock ResponseVO对象，再mock其方法',
        target_section='Mockito异常处理经验'
    )
    
    # 结束会话，显示总结
    print("\n\n=== 结束会话 ===")
    tracker.end_session()


import json
import os
from datetime import datetime
from typing import List, Dict, Optional

class ExperienceTracker:
    """经验积累追踪器类"""
    
    def __init__(self, skill_dir: str = '.'):
        """
        初始化追踪器
        
        Args:
            skill_dir: skill目录路径
        """
        self.skill_dir = skill_dir
        self.tracker_file = os.path.join(skill_dir, '.experience_tracker.json')
        self.guide_file = os.path.join(skill_dir, 'references/troubleshooting-guide.md')
        self.load_tracker()
    
    def load_tracker(self):
        """加载追踪状态"""
        if os.path.exists(self.tracker_file):
            with open(self.tracker_file, 'r', encoding='utf-8') as f:
                self.data = json.load(f)
        else:
            self.data = {
                'current_session': None,
                'pending_experiences': [],
                'last_guide_update': None,
                'session_history': []
            }
    
    def save_tracker(self):
        """保存追踪状态"""
        with open(self.tracker_file, 'w', encoding='utf-8') as f:
            json.dump(self.data, f, ensure_ascii=False, indent=2)
    
    def start_session(self, test_class_name: str) -> bool:
        """
        开始新的测试会话
        
        Args:
            test_class_name: 测试类名
            
        Returns:
            bool: 是否成功启动
        """
        # 检查是否有未完成的会话
        if self.data['current_session'] and self.data['pending_experiences']:
            print("\n" + "⚠️ "*30)
            print("警告：存在未完成的测试会话！")
            print("⚠️ "*30)
            current = self.data['current_session']
            print(f"未完成的会话：{current['test_class']}")
            print(f"待记录经验数：{len(self.data['pending_experiences'])}")
            print("\n是否要继续未完成的会话？(y/n)")
            # 在实际使用中，这里应该由AI或用户决定
            return False
        
        self.data['current_session'] = {
            'test_class': test_class_name,
            'start_time': datetime.now().isoformat(),
            'problems_fixed': 0,
            'problems_recorded': 0
        }
        self.data['pending_experiences'] = []
        self.save_tracker()
        
        print(f"\n✅ 已启动测试会话: {test_class_name}")
        print(f"开始时间: {self.data['current_session']['start_time']}")
        print("-" * 60 + "\n")
        return True
    
    def record_problem_fixed(self, problem_type: str, description: str, 
                            error_message: str = "") -> bool:
        """
        记录已修复的问题（等待经验积累）
        
        Args:
            problem_type: 问题类型 ('compile_error' | 'test_failure' | 'framework_issue' | 'mock_config')
            description: 问题描述
            error_message: 错误信息（可选）
            
        Returns:
            bool: 是否允许继续（False表示必须先记录经验）
        """
        problem = {
            'type': problem_type,
            'description': description,
            'error_message': error_message,
            'fixed_time': datetime.now().isoformat(),
            'recorded': False
        }
        self.data['pending_experiences'].append(problem)
        self.data['current_session']['problems_fixed'] += 1
        self.save_tracker()
        
        # 🔴 强制提示：必须立即积累经验
        self._show_mandatory_prompt(problem)
        
        # 返回 False 表示不允许继续
        return False
    
    def _show_mandatory_prompt(self, problem: Dict):
        """显示强制记录提示"""
        print("\n" + "="*80)
        print("🔴 强制检查点：问题已修复，但经验尚未记录！")
        print("="*80)
        print(f"\n📋 问题信息：")
        print(f"   类型: {self._get_problem_type_name(problem['type'])}")
        print(f"   描述: {problem['description']}")
        if problem['error_message']:
            print(f"   错误: {problem['error_message'][:100]}...")
        print(f"   修复时间: {problem['fixed_time']}")
        
        print(f"\n📝 记录位置：")
        print(f"   文件: {self.guide_file}")
        print(f"   章节: {self._get_recommended_section(problem['type'])}")
        
        print("\n⛔ 强制要求：")
        print("   1. 立即打开 troubleshooting-guide.md")
        print("   2. 在推荐章节记录此问题")
        print("   3. 包含：问题类型、错误信息、根本原因、解决方案")
        print("   4. 如果是测试失败，需包含代码示例（✅正确 ❌错误）")
        print("   5. 保存文件")
        print("   6. 调用 mark_experience_recorded() 标记已完成")
        
        print("\n" + "⛔"*40)
        print("在记录完成前，不允许继续执行测试！")
        print("⛔"*40 + "\n")
    
    def _get_problem_type_name(self, problem_type: str) -> str:
        """获取问题类型的中文名称"""
        type_names = {
            'compile_error': '编译错误',
            'test_failure': '测试失败',
            'framework_issue': '框架问题',
            'mock_config': 'Mock配置问题'
        }
        return type_names.get(problem_type, problem_type)
    
    def _get_recommended_section(self, problem_type: str) -> str:
        """获取推荐的记录章节"""
        sections = {
            'compile_error': '常见问题速查表',
            'test_failure': 'Mockito异常处理经验 或 最新经验积累',
            'framework_issue': '常见问题速查表',
            'mock_config': 'Mockito异常处理经验'
        }
        return sections.get(problem_type, '常见问题速查表')
    
    def mark_experience_recorded(self, problem_index: int = -1) -> bool:
        """
        标记经验已记录
        
        Args:
            problem_index: 要标记的问题索引，默认-1表示最后一个
            
        Returns:
            bool: 是否标记成功
        """
        if not self.data['pending_experiences']:
            print("✅ 没有待记录的经验")
            return True
        
        # 检查文件是否真的被修改了
        if not self._verify_guide_updated():
            print("\n" + "❌"*40)
            print("错误: troubleshooting-guide.md 文件未被修改！")
            print("❌"*40)
            print("\n请确认：")
            print("  1. 已打开并编辑 troubleshooting-guide.md")
            print("  2. 已添加问题记录到相应章节")
            print("  3. 已保存文件")
            print("\n完成后再次调用 mark_experience_recorded()")
            print("="*80 + "\n")
            return False
        
        # 标记问题为已记录
        problem = self.data['pending_experiences'][problem_index]
        problem['recorded'] = True
        problem['recorded_time'] = datetime.now().isoformat()
        
        self.data['current_session']['problems_recorded'] += 1
        self.data['last_guide_update'] = datetime.now().isoformat()
        self.save_tracker()
        
        print("\n" + "✅"*40)
        print(f"经验已标记为已记录")
        print("✅"*40)
        print(f"\n记录的问题: {problem['description']}")
        print(f"记录时间: {problem['recorded_time']}")
        print(f"\n当前进度: {self.data['current_session']['problems_recorded']}/{self.data['current_session']['problems_fixed']} 已记录")
        print("="*80 + "\n")
        
        return True
    
    def _verify_guide_updated(self) -> bool:
        """
        验证 troubleshooting-guide.md 是否在最近被更新
        
        Returns:
            bool: 文件是否被更新
        """
        if not os.path.exists(self.guide_file):
            return False
        
        # 检查文件修改时间
        mtime = os.path.getmtime(self.guide_file)
        last_update = self.data.get('last_guide_update')
        
        if last_update:
            last_update_time = datetime.fromisoformat(last_update).timestamp()
            # 文件修改时间必须晚于上次记录时间
            return mtime > last_update_time
        
        # 如果是第一次记录，检查文件是否在最近1小时内被修改
        now = datetime.now().timestamp()
        return (now - mtime) < 3600  # 1小时
    
    def check_before_continue(self) -> bool:
        """
        在继续下一步前检查是否有未记录的经验
        
        Returns:
            bool: 是否允许继续
        """
        pending = [p for p in self.data['pending_experiences'] if not p['recorded']]
        
        if not pending:
            print("✅ 所有经验已记录，可以继续")
            return True
        
        # 有未记录的经验，显示错误
        print("\n" + "🛑"*40)
        print("错误：存在未记录的经验，不允许继续执行！")
        print("🛑"*40)
        print(f"\n待记录的问题数量: {len(pending)}")
        
        for i, p in enumerate(pending, 1):
            print(f"\n问题 {i}:")
            print(f"  类型: {self._get_problem_type_name(p['type'])}")
            print(f"  描述: {p['description']}")
            print(f"  修复时间: {p['fixed_time']}")
        
        print("\n⛔ 必须先完成以下操作才能继续:")
        print("   1. 打开 troubleshooting-guide.md")
        print("   2. 记录所有待记录的经验到相应章节")
        print("   3. 保存文件")
        print("   4. 调用 mark_experience_recorded() 标记已完成")
        print("\n" + "🛑"*40 + "\n")
        
        return False
    
    def end_session(self) -> Dict:
        """
        结束当前测试会话
        
        Returns:
            Dict: 会话总结信息
        """
        if not self.data['current_session']:
            return {'error': '无活跃会话'}
        
        # 检查是否有未记录的经验
        if not self.check_before_continue():
            return {'error': '存在未记录的经验，不能结束会话'}
        
        # 生成会话总结
        session = self.data['current_session']
        summary = {
            'test_class': session['test_class'],
            'start_time': session['start_time'],
            'end_time': datetime.now().isoformat(),
            'problems_fixed': session['problems_fixed'],
            'problems_recorded': session['problems_recorded'],
            'completion_rate': f"{session['problems_recorded']}/{session['problems_fixed']}",
            'experiences': self.data['pending_experiences']
        }
        
        # 保存到历史记录
        self.data['session_history'].append(summary)
        
        # 清空当前会话
        self.data['current_session'] = None
        self.data['pending_experiences'] = []
        self.save_tracker()
        
        return summary
    
    def generate_summary(self) -> str:
        """
        生成本次会话的经验积累总结
        
        Returns:
            str: 格式化的总结文本
        """
        if not self.data['current_session']:
            return "❌ 无活跃会话"
        
        session = self.data['current_session']
        pending = [p for p in self.data['pending_experiences'] if not p['recorded']]
        recorded = [p for p in self.data['pending_experiences'] if p['recorded']]
        
        summary = f"""
{'='*80}
📊 单元测试会话总结
{'='*80}

📌 测试类: {session['test_class']}
⏰ 开始时间: {session['start_time']}
⏱️  当前时间: {datetime.now().isoformat()}

📈 问题修复统计:
   - 修复问题总数: {session['problems_fixed']}
   - 已记录经验数: {session['problems_recorded']}
   - 待记录经验数: {len(pending)}
   - 积累完成率: {session['problems_recorded']}/{session['problems_fixed']} ({session['problems_recorded']/session['problems_fixed']*100:.1f}%)

"""
        
        if recorded:
            summary += "✅ 已记录的经验:\n"
            for i, p in enumerate(recorded, 1):
                summary += f"   {i}. [{self._get_problem_type_name(p['type'])}] {p['description']}\n"
            summary += "\n"
        
        if pending:
            summary += "⚠️  待记录的经验:\n"
            for i, p in enumerate(pending, 1):
                summary += f"   {i}. [{self._get_problem_type_name(p['type'])}] {p['description']}\n"
            summary += "\n"
        
        summary += "="*80
        
        return summary
    
    def get_status(self) -> Dict:
        """
        获取当前追踪状态
        
        Returns:
            Dict: 状态信息
        """
        if not self.data['current_session']:
            return {'status': 'no_session', 'message': '无活跃会话'}
        
        pending = [p for p in self.data['pending_experiences'] if not p['recorded']]
        
        return {
            'status': 'active',
            'test_class': self.data['current_session']['test_class'],
            'problems_fixed': self.data['current_session']['problems_fixed'],
            'problems_recorded': self.data['current_session']['problems_recorded'],
            'pending_count': len(pending),
            'can_continue': len(pending) == 0
        }


# 使用示例
if __name__ == '__main__':
    # 初始化追踪器
    tracker = ExperienceTracker('.')
    
    # 示例：开始测试会话
    tracker.start_session('CadreDevHandInAuditerServiceImplTest')
    
    # 示例：修复编译错误
    print("模拟场景：修复编译错误...")
    tracker.record_problem_fixed(
        'compile_error',
        '静态Mock不支持 - 缺少mockito-inline依赖',
        'MockMaker does not support static mocks'
    )
    
    # 此时会显示强制提示，不允许继续
    # 用户需要：
    # 1. 打开 troubleshooting-guide.md
    # 2. 记录问题到「常见问题速查表」
    # 3. 保存文件
    # 4. 调用 tracker.mark_experience_recorded()
    
    print("\n\n模拟场景：用户已记录经验...")
    # 注意：在实际使用中，需要等用户真的修改了文件
    tracker.mark_experience_recorded()
    
    # 检查是否可以继续
    if tracker.check_before_continue():
        print("✅ 可以继续执行下一步")
    
    # 生成总结
    print(tracker.generate_summary())
