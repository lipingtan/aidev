/**
 * 设备类型检测工具
 * 运行时判断当前访问设备是否为移动端，用于自动切换 PC/H5 布局
 */

/**
 * 检测当前设备是否为移动端（手机/平板）
 * 综合 User-Agent 和屏幕宽度双重判断
 */
export function isMobile(): boolean {
  // User-Agent 检测
  const uaMatch = /Android|iPhone|iPad|iPod|Mobile|IEMobile|Opera Mini|BlackBerry|Windows Phone/i
    .test(navigator.userAgent)
  // 屏幕宽度检测（宽度 < 768px 视为移动端）
  const widthMatch = window.innerWidth < 768
  return uaMatch || widthMatch
}
