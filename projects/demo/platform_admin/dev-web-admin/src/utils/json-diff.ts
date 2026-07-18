/**
 * JSON Diff 工具
 * 使用 jsondiffpatch 生成 HTML 格式的 diff 展示
 */
import { create, formatters } from 'jsondiffpatch'

const diffInstance = create({
  objectHash: (obj: any) => obj.id || JSON.stringify(obj)
})

/**
 * 生成 diff HTML
 * @param oldVal 旧值
 * @param newVal 新值
 * @returns HTML 字符串，包含差异高亮
 */
export function formatDiffHtml(oldVal: any, newVal: any): string {
  if (!oldVal && !newVal) return '<span class="no-diff">无数据</span>'
  const delta = diffInstance.diff(oldVal || {}, newVal || {})
  if (!delta) return '<span class="no-diff">无变更</span>'
  return formatters.html.format(delta, oldVal || {})
}
