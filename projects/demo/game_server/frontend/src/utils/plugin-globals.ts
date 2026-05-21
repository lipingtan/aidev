// 将 Vue 和 Element Plus 挂载到 window，供插件 bundle 使用
import * as Vue from "vue";
import * as ElementPlus from "element-plus";

(window as any).__HOST_VUE__ = Vue;
(window as any).__HOST_ELEMENT_PLUS__ = ElementPlus;
