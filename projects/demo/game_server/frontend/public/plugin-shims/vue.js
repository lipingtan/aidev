// 从 Host 挂载的全局变量重新导出所有 Vue API
const Vue = window.__HOST_VUE__;
// 逐个导出所有 Vue 公开 API
export const {
  ref, reactive, computed, watch, watchEffect,
  onMounted, onUnmounted, onBeforeMount, onBeforeUnmount, onUpdated, onBeforeUpdate,
  defineComponent, h, createApp, nextTick,
  shallowRef, shallowReactive, triggerRef, toRaw, markRaw,
  toRefs, toRef, unref, isRef, isReactive, isReadonly, isProxy,
  provide, inject, getCurrentInstance,
  createElementVNode, createVNode, createTextVNode, createCommentVNode, createStaticVNode,
  openBlock, createBlock, createElementBlock,
  Fragment, Teleport, Suspense, KeepAlive, Transition, TransitionGroup,
  resolveComponent, resolveDirective, resolveDynamicComponent,
  withDirectives, withModifiers, withCtx, withKeys,
  renderList, renderSlot, normalizeClass, normalizeStyle, normalizeProps,
  guardReactiveProps, mergeProps, cloneVNode,
  toDisplayString, camelize, capitalize,
  defineEmits, defineProps, defineExpose, defineOptions, withDefaults,
  useSlots, useAttrs,
  effectScope, onScopeDispose,
  customRef, readonly, shallowReadonly,
  toHandlers, toHandlerKey,
  vShow, vModelText, vModelCheckbox, vModelRadio, vModelSelect, vModelDynamic,
  pushScopeId, popScopeId, withScopeId,
  setBlockTracking, createPropsRestProxy,
  useCssVars, useCssModule,
  ssrContextKey, useSSRContext,
  initCustomFormatter, warn,
  version, compile
} = Vue;
export default Vue;
