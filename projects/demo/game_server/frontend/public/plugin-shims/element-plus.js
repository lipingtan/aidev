// 从 Host 挂载的全局变量重新导出 Element Plus
const EP = window.__HOST_ELEMENT_PLUS__;
export const ElMessage = EP.ElMessage;
export const ElMessageBox = EP.ElMessageBox;
export const ElNotification = EP.ElNotification;
export const ElLoading = EP.ElLoading;
export default EP;
