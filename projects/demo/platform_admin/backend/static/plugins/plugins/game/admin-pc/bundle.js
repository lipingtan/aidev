import { defineComponent as te, ref as I, reactive as H, onMounted as le, resolveComponent as s, resolveDirective as ae, openBlock as P, createElementBlock as Z, createElementVNode as j, createVNode as e, withCtx as t, createTextVNode as u, withDirectives as ne, createBlock as M, toDisplayString as R, watch as $e, Fragment as ie, renderList as ve, createCommentVNode as Q } from "/plugin-shims/vue.js";
import { ElMessage as z } from "/plugin-shims/element-plus.js";
const De = { class: "plugin-page" }, ze = { class: "page-header" }, de = "/api/v1/plugin/game/game", Pe = /* @__PURE__ */ te({
  __name: "GameList",
  setup(h) {
    const L = I(!1), T = I(!1), E = I([]), k = I(0), $ = I(!1), q = I(), y = H({ name: "", pageIndex: 1, pageSize: 10 }), g = H({ id: null, name: "", version: "1.0.0", description: "", status: 1 }), i = {
      name: [{ required: !0, message: "请输入游戏名称", trigger: "blur" }]
    };
    function V() {
      try {
        const a = document.cookie.match(/authorized-token=([^;]+)/);
        if (a)
          return JSON.parse(decodeURIComponent(a[1]))?.accessToken || "";
        const l = localStorage.getItem("responsive-user-info");
        if (l)
          return JSON.parse(l)?.accessToken || "";
      } catch {
      }
      return "";
    }
    async function b(a, l, v) {
      const d = {
        Authorization: `Bearer ${V()}`
      }, o = { method: a, headers: d };
      return v && (d["Content-Type"] = "application/json", o.body = JSON.stringify(v)), (await fetch(l, o)).json();
    }
    async function O() {
      L.value = !0;
      try {
        const a = new URLSearchParams();
        a.set("page", String(y.pageIndex)), a.set("pageSize", String(y.pageSize)), y.name && a.set("name", y.name);
        const l = await b("GET", `${de}?${a.toString()}`);
        l.code === 200 && (E.value = l.data?.list || [], k.value = l.data?.total || 0);
      } finally {
        L.value = !1;
      }
    }
    function K() {
      y.pageIndex = 1, O();
    }
    function w() {
      Object.assign(y, { name: "", pageIndex: 1 }), O();
    }
    function p(a) {
      Object.assign(g, { id: null, name: "", version: "1.0.0", description: "", status: 1 }), a && Object.assign(g, a), $.value = !0;
    }
    async function F() {
      await q.value?.validate(), T.value = !0;
      try {
        const a = g.id ? await b("PUT", `${de}/${g.id}`, g) : await b("POST", de, g);
        a.code === 200 ? (z.success("操作成功"), $.value = !1, O()) : z.error(a.msg || "操作失败");
      } finally {
        T.value = !1;
      }
    }
    async function N(a) {
      const l = await b("DELETE", `${de}/${a.id}`);
      l.code === 200 ? (z.success("删除成功"), O()) : z.error(l.msg || "删除失败");
    }
    async function x(a) {
      const l = await b("POST", `${de}/${a.id}/regen-secret`);
      l.code === 200 ? (z.success("密钥已重置"), O()) : z.error(l.msg || "操作失败");
    }
    return le(O), (a, l) => {
      const v = s("el-button"), d = s("el-input"), o = s("el-form-item"), D = s("el-form"), U = s("el-table-column"), J = s("el-tag"), W = s("el-popconfirm"), S = s("el-table"), B = s("el-pagination"), A = s("el-radio"), f = s("el-radio-group"), n = s("el-dialog"), r = ae("loading");
      return P(), Z("div", De, [
        j("div", ze, [
          l[11] || (l[11] = j("h3", null, "游戏管理", -1)),
          j("div", null, [
            e(v, {
              type: "primary",
              onClick: l[0] || (l[0] = (_) => p())
            }, {
              default: t(() => [...l[10] || (l[10] = [
                u("新增游戏", -1)
              ])]),
              _: 1
            })
          ])
        ]),
        e(D, {
          inline: !0,
          model: y,
          style: { "margin-bottom": "16px" }
        }, {
          default: t(() => [
            e(o, { label: "游戏名称" }, {
              default: t(() => [
                e(d, {
                  modelValue: y.name,
                  "onUpdate:modelValue": l[1] || (l[1] = (_) => y.name = _),
                  placeholder: "请输入游戏名称",
                  clearable: ""
                }, null, 8, ["modelValue"])
              ]),
              _: 1
            }),
            e(o, null, {
              default: t(() => [
                e(v, {
                  type: "primary",
                  onClick: K
                }, {
                  default: t(() => [...l[12] || (l[12] = [
                    u("查询", -1)
                  ])]),
                  _: 1
                }),
                e(v, { onClick: w }, {
                  default: t(() => [...l[13] || (l[13] = [
                    u("重置", -1)
                  ])]),
                  _: 1
                })
              ]),
              _: 1
            })
          ]),
          _: 1
        }, 8, ["model"]),
        ne((P(), M(S, {
          data: E.value,
          stripe: ""
        }, {
          default: t(() => [
            e(U, {
              prop: "id",
              label: "ID",
              width: "60"
            }),
            e(U, {
              prop: "name",
              label: "游戏名称"
            }),
            e(U, {
              prop: "appKey",
              label: "AppKey",
              width: "120"
            }),
            e(U, {
              prop: "version",
              label: "版本",
              width: "80"
            }),
            e(U, {
              label: "状态",
              width: "80"
            }, {
              default: t(({ row: _ }) => [
                e(J, {
                  type: _.status === 1 ? "success" : "danger"
                }, {
                  default: t(() => [
                    u(R(_.status === 1 ? "启用" : "禁用"), 1)
                  ]),
                  _: 2
                }, 1032, ["type"])
              ]),
              _: 1
            }),
            e(U, {
              prop: "createdAt",
              label: "创建时间",
              width: "160"
            }),
            e(U, {
              label: "操作",
              width: "200",
              fixed: "right"
            }, {
              default: t(({ row: _ }) => [
                e(v, {
                  link: "",
                  type: "primary",
                  onClick: (m) => p(_)
                }, {
                  default: t(() => [...l[14] || (l[14] = [
                    u("编辑", -1)
                  ])]),
                  _: 1
                }, 8, ["onClick"]),
                e(v, {
                  link: "",
                  type: "warning",
                  onClick: (m) => x(_)
                }, {
                  default: t(() => [...l[15] || (l[15] = [
                    u("重置密钥", -1)
                  ])]),
                  _: 1
                }, 8, ["onClick"]),
                e(W, {
                  title: "确认删除？",
                  onConfirm: (m) => N(_)
                }, {
                  reference: t(() => [
                    e(v, {
                      link: "",
                      type: "danger"
                    }, {
                      default: t(() => [...l[16] || (l[16] = [
                        u("删除", -1)
                      ])]),
                      _: 1
                    })
                  ]),
                  _: 1
                }, 8, ["onConfirm"])
              ]),
              _: 1
            })
          ]),
          _: 1
        }, 8, ["data"])), [
          [r, L.value]
        ]),
        e(B, {
          style: { "margin-top": "16px", display: "flex", "justify-content": "flex-end" },
          "current-page": y.pageIndex,
          "onUpdate:currentPage": l[2] || (l[2] = (_) => y.pageIndex = _),
          "page-size": y.pageSize,
          "onUpdate:pageSize": l[3] || (l[3] = (_) => y.pageSize = _),
          total: k.value,
          layout: "total, sizes, prev, pager, next",
          onChange: O
        }, null, 8, ["current-page", "page-size", "total"]),
        e(n, {
          modelValue: $.value,
          "onUpdate:modelValue": l[9] || (l[9] = (_) => $.value = _),
          title: g.id ? "编辑游戏" : "新增游戏",
          width: "500px"
        }, {
          footer: t(() => [
            e(v, {
              onClick: l[8] || (l[8] = (_) => $.value = !1)
            }, {
              default: t(() => [...l[19] || (l[19] = [
                u("取消", -1)
              ])]),
              _: 1
            }),
            e(v, {
              type: "primary",
              loading: T.value,
              onClick: F
            }, {
              default: t(() => [...l[20] || (l[20] = [
                u("确定", -1)
              ])]),
              _: 1
            }, 8, ["loading"])
          ]),
          default: t(() => [
            e(D, {
              ref_key: "formRef",
              ref: q,
              model: g,
              rules: i,
              "label-width": "90px"
            }, {
              default: t(() => [
                e(o, {
                  label: "游戏名称",
                  prop: "name"
                }, {
                  default: t(() => [
                    e(d, {
                      modelValue: g.name,
                      "onUpdate:modelValue": l[4] || (l[4] = (_) => g.name = _)
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(o, { label: "版本" }, {
                  default: t(() => [
                    e(d, {
                      modelValue: g.version,
                      "onUpdate:modelValue": l[5] || (l[5] = (_) => g.version = _),
                      placeholder: "1.0.0"
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(o, { label: "描述" }, {
                  default: t(() => [
                    e(d, {
                      modelValue: g.description,
                      "onUpdate:modelValue": l[6] || (l[6] = (_) => g.description = _),
                      type: "textarea"
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(o, { label: "状态" }, {
                  default: t(() => [
                    e(f, {
                      modelValue: g.status,
                      "onUpdate:modelValue": l[7] || (l[7] = (_) => g.status = _)
                    }, {
                      default: t(() => [
                        e(A, { value: 1 }, {
                          default: t(() => [...l[17] || (l[17] = [
                            u("启用", -1)
                          ])]),
                          _: 1
                        }),
                        e(A, { value: 2 }, {
                          default: t(() => [...l[18] || (l[18] = [
                            u("禁用", -1)
                          ])]),
                          _: 1
                        })
                      ]),
                      _: 1
                    }, 8, ["modelValue"])
                  ]),
                  _: 1
                })
              ]),
              _: 1
            }, 8, ["model"])
          ]),
          _: 1
        }, 8, ["modelValue", "title"])
      ]);
    };
  }
}), oe = (h, L) => {
  const T = h.__vccOpts || h;
  for (const [E, k] of L)
    T[E] = k;
  return T;
}, Te = /* @__PURE__ */ oe(Pe, [["__scopeId", "data-v-c2cb1ebd"]]), Re = { class: "plugin-page" }, Le = { style: { "margin-bottom": "16px" } }, Oe = { style: { display: "flex", gap: "16px", "margin-bottom": "16px" } }, ee = "/api/v1/plugin/game/dlc", Ne = "/api/v1/plugin/game/game", Ke = 500 * 1024 * 1024, Ae = /* @__PURE__ */ te({
  __name: "DlcList",
  setup(h) {
    const L = I("list"), T = I(!1), E = I(!1), k = I(!1), $ = I(!1), q = I(), y = I([]), g = I(0), i = I([]), V = H({ gameId: "", name: "", pageIndex: 1, pageSize: 10 }), b = H({ id: null, gameId: "", dlcKey: "", name: "", version: "", description: "", price: 0, isFree: 1, status: 1, minGameVersion: "" }), O = {
      gameId: [{ required: !0, message: "请选择游戏", trigger: "change" }],
      dlcKey: [{ required: !0, message: "请输入DLC Key", trigger: "blur" }],
      name: [{ required: !0, message: "请输入DLC名称", trigger: "blur" }],
      version: [{ required: !0, message: "请输入版本", trigger: "blur" }],
      price: [{ required: !0, message: "请输入价格", trigger: "blur" }]
    }, K = I(null), w = I([]), p = H({ totalDownloads: 0, totalRevenue: 0, ranking: [] });
    function F() {
      try {
        const f = document.cookie.match(/authorized-token=([^;]+)/);
        if (f)
          return JSON.parse(decodeURIComponent(f[1]))?.accessToken || "";
        const n = localStorage.getItem("responsive-user-info");
        if (n)
          return JSON.parse(n)?.accessToken || "";
      } catch {
      }
      return "";
    }
    async function N(f, n, r) {
      const _ = {
        Authorization: `Bearer ${F()}`
      }, m = { method: f, headers: _ };
      return r && (_["Content-Type"] = "application/json", m.body = JSON.stringify(r)), (await fetch(n, m)).json();
    }
    function x(f) {
      return f < 1024 ? f + " B" : f < 1024 * 1024 ? (f / 1024).toFixed(1) + " KB" : (f / (1024 * 1024)).toFixed(1) + " MB";
    }
    async function a() {
      T.value = !0;
      try {
        const f = new URLSearchParams();
        f.set("pageIndex", String(V.pageIndex)), f.set("pageSize", String(V.pageSize)), V.gameId && f.set("gameId", String(V.gameId)), V.name && f.set("name", V.name);
        const n = await N("GET", `${ee}?${f.toString()}`);
        n.code === 200 && (y.value = n.data?.list || n.data || [], g.value = n.data?.count || n.count || 0);
      } finally {
        T.value = !1;
      }
    }
    async function l() {
      try {
        const f = await N("GET", `${Ne}?pageSize=100`);
        f.code === 200 && (i.value = f.data?.list || f.data || []);
      } catch {
        i.value = [];
      }
    }
    async function v() {
      try {
        const f = V.gameId ? `?gameId=${V.gameId}` : "", n = await N("GET", `${ee}/stats${f}`);
        if (n.code === 200) {
          const r = n.data || {};
          p.totalDownloads = r.totalDownloads || 0, p.totalRevenue = r.totalRevenue || 0, p.ranking = r.ranking || [];
        }
      } catch {
      }
    }
    function d() {
      V.pageIndex = 1, a();
    }
    function o() {
      Object.assign(V, { gameId: "", name: "", pageIndex: 1 }), a();
    }
    function D(f) {
      f === "stats" && v();
    }
    function U(f) {
      Object.assign(b, { id: null, gameId: "", dlcKey: "", name: "", version: "", description: "", price: 0, isFree: 1, status: 1, minGameVersion: "" }), K.value = null, w.value = [], f && (Object.assign(b, f), b.price = f.price / 100), $.value = !0;
    }
    async function J() {
      await q.value?.validate(), E.value = !0;
      try {
        const f = { ...b, price: Math.round(b.price * 100) }, n = b.id ? await N("PUT", `${ee}/${b.id}`, f) : await N("POST", ee, f);
        n.code === 200 ? (z.success("操作成功"), $.value = !1, a()) : z.error(n.msg || "操作失败");
      } finally {
        E.value = !1;
      }
    }
    async function W(f) {
      const n = await N("DELETE", `${ee}/${f.id}`);
      n.code === 200 ? (z.success("删除成功"), a()) : z.error(n.msg || "删除失败");
    }
    async function S(f) {
      try {
        const n = await N("GET", `${ee}/${f.id}/download-url`);
        if (n.code === 200) {
          const r = n.data?.url || n.url;
          r ? window.open(r) : z.error("获取下载地址失败");
        } else
          z.error(n.msg || "获取下载地址失败");
      } catch {
        z.error("获取下载地址失败");
      }
    }
    function B(f) {
      if (f.raw && f.raw.size > Ke) {
        z.error("文件大小不能超过 500MB"), w.value = [], K.value = null;
        return;
      }
      K.value = f.raw || null, w.value = f.raw ? [f] : [];
    }
    async function A() {
      if (!(!K.value || !b.id)) {
        k.value = !0;
        try {
          const f = new FormData();
          f.append("file", K.value);
          const r = await (await fetch(`${ee}/${b.id}/upload`, {
            method: "POST",
            headers: { Authorization: `Bearer ${F()}` },
            body: f
          })).json();
          r.code === 200 ? (z.success("上传成功"), K.value = null, w.value = [], a()) : z.error(r.msg || "上传失败");
        } catch {
          z.error("上传失败");
        } finally {
          k.value = !1;
        }
      }
    }
    return $e(() => b.isFree, (f) => {
      f === 1 && (b.price = 0);
    }), le(() => {
      l(), a();
    }), (f, n) => {
      const r = s("el-option"), _ = s("el-select"), m = s("el-form-item"), X = s("el-input"), Y = s("el-button"), C = s("el-form"), G = s("el-table-column"), me = s("el-tag"), Ie = s("el-popconfirm"), fe = s("el-table"), ke = s("el-pagination"), ge = s("el-tab-pane"), ce = s("el-statistic"), be = s("el-card"), we = s("el-tabs"), se = s("el-radio"), ye = s("el-radio-group"), xe = s("el-input-number"), Se = s("el-upload"), Ue = s("el-dialog"), Ce = ae("loading");
      return P(), Z("div", Re, [
        n[32] || (n[32] = j("div", { class: "page-header" }, [
          j("h3", null, "DLC管理"),
          j("div")
        ], -1)),
        e(we, {
          modelValue: L.value,
          "onUpdate:modelValue": n[5] || (n[5] = (c) => L.value = c),
          onTabChange: D
        }, {
          default: t(() => [
            e(ge, {
              label: "列表管理",
              name: "list"
            }, {
              default: t(() => [
                e(C, {
                  inline: !0,
                  model: V,
                  style: { "margin-bottom": "16px" }
                }, {
                  default: t(() => [
                    e(m, { label: "游戏" }, {
                      default: t(() => [
                        e(_, {
                          modelValue: V.gameId,
                          "onUpdate:modelValue": n[0] || (n[0] = (c) => V.gameId = c),
                          placeholder: "请选择游戏",
                          clearable: "",
                          style: { width: "200px" }
                        }, {
                          default: t(() => [
                            (P(!0), Z(ie, null, ve(i.value, (c) => (P(), M(r, {
                              key: c.id,
                              label: c.name,
                              value: c.id
                            }, null, 8, ["label", "value"]))), 128))
                          ]),
                          _: 1
                        }, 8, ["modelValue"])
                      ]),
                      _: 1
                    }),
                    e(m, { label: "DLC名称" }, {
                      default: t(() => [
                        e(X, {
                          modelValue: V.name,
                          "onUpdate:modelValue": n[1] || (n[1] = (c) => V.name = c),
                          placeholder: "请输入DLC名称",
                          clearable: ""
                        }, null, 8, ["modelValue"])
                      ]),
                      _: 1
                    }),
                    e(m, null, {
                      default: t(() => [
                        e(Y, {
                          type: "primary",
                          onClick: d
                        }, {
                          default: t(() => [...n[17] || (n[17] = [
                            u("查询", -1)
                          ])]),
                          _: 1
                        }),
                        e(Y, { onClick: o }, {
                          default: t(() => [...n[18] || (n[18] = [
                            u("重置", -1)
                          ])]),
                          _: 1
                        })
                      ]),
                      _: 1
                    })
                  ]),
                  _: 1
                }, 8, ["model"]),
                j("div", Le, [
                  e(Y, {
                    type: "primary",
                    onClick: n[2] || (n[2] = (c) => U())
                  }, {
                    default: t(() => [...n[19] || (n[19] = [
                      u("新增DLC", -1)
                    ])]),
                    _: 1
                  })
                ]),
                ne((P(), M(fe, {
                  data: y.value,
                  stripe: ""
                }, {
                  default: t(() => [
                    e(G, {
                      prop: "id",
                      label: "ID",
                      width: "60"
                    }),
                    e(G, {
                      prop: "gameId",
                      label: "游戏ID",
                      width: "80"
                    }),
                    e(G, {
                      prop: "name",
                      label: "DLC名称"
                    }),
                    e(G, {
                      prop: "dlcKey",
                      label: "DLC Key"
                    }),
                    e(G, {
                      prop: "version",
                      label: "版本",
                      width: "80"
                    }),
                    e(G, {
                      label: "价格(元)",
                      width: "100"
                    }, {
                      default: t(({ row: c }) => [
                        u(R((c.price / 100).toFixed(2)), 1)
                      ]),
                      _: 1
                    }),
                    e(G, {
                      label: "是否免费",
                      width: "90"
                    }, {
                      default: t(({ row: c }) => [
                        e(me, {
                          type: c.isFree === 1 ? "success" : "warning"
                        }, {
                          default: t(() => [
                            u(R(c.isFree === 1 ? "免费" : "付费"), 1)
                          ]),
                          _: 2
                        }, 1032, ["type"])
                      ]),
                      _: 1
                    }),
                    e(G, {
                      label: "文件大小",
                      width: "100"
                    }, {
                      default: t(({ row: c }) => [
                        u(R(c.fileSize > 0 ? x(c.fileSize) : "-"), 1)
                      ]),
                      _: 1
                    }),
                    e(G, {
                      label: "状态",
                      width: "80"
                    }, {
                      default: t(({ row: c }) => [
                        e(me, {
                          type: c.status === 1 ? "success" : "danger"
                        }, {
                          default: t(() => [
                            u(R(c.status === 1 ? "上架" : "下架"), 1)
                          ]),
                          _: 2
                        }, 1032, ["type"])
                      ]),
                      _: 1
                    }),
                    e(G, {
                      prop: "downloadCount",
                      label: "下载次数",
                      width: "90"
                    }),
                    e(G, {
                      prop: "createdAt",
                      label: "创建时间",
                      width: "160"
                    }),
                    e(G, {
                      label: "操作",
                      width: "220",
                      fixed: "right"
                    }, {
                      default: t(({ row: c }) => [
                        e(Y, {
                          link: "",
                          type: "primary",
                          onClick: (_e) => U(c)
                        }, {
                          default: t(() => [...n[20] || (n[20] = [
                            u("编辑", -1)
                          ])]),
                          _: 1
                        }, 8, ["onClick"]),
                        c.filePath ? (P(), M(Y, {
                          key: 0,
                          link: "",
                          type: "success",
                          onClick: (_e) => S(c)
                        }, {
                          default: t(() => [...n[21] || (n[21] = [
                            u("下载", -1)
                          ])]),
                          _: 1
                        }, 8, ["onClick"])) : Q("", !0),
                        e(Ie, {
                          title: "确认删除？",
                          onConfirm: (_e) => W(c)
                        }, {
                          reference: t(() => [
                            e(Y, {
                              link: "",
                              type: "danger"
                            }, {
                              default: t(() => [...n[22] || (n[22] = [
                                u("删除", -1)
                              ])]),
                              _: 1
                            })
                          ]),
                          _: 1
                        }, 8, ["onConfirm"])
                      ]),
                      _: 1
                    })
                  ]),
                  _: 1
                }, 8, ["data"])), [
                  [Ce, T.value]
                ]),
                e(ke, {
                  style: { "margin-top": "16px", display: "flex", "justify-content": "flex-end" },
                  "current-page": V.pageIndex,
                  "onUpdate:currentPage": n[3] || (n[3] = (c) => V.pageIndex = c),
                  "page-size": V.pageSize,
                  "onUpdate:pageSize": n[4] || (n[4] = (c) => V.pageSize = c),
                  total: g.value,
                  layout: "total, sizes, prev, pager, next",
                  onChange: a
                }, null, 8, ["current-page", "page-size", "total"])
              ]),
              _: 1
            }),
            e(ge, {
              label: "统计分析",
              name: "stats"
            }, {
              default: t(() => [
                j("div", Oe, [
                  e(be, {
                    shadow: "hover",
                    style: { width: "240px" }
                  }, {
                    default: t(() => [
                      e(ce, {
                        title: "总下载次数",
                        value: p.totalDownloads
                      }, null, 8, ["value"])
                    ]),
                    _: 1
                  }),
                  e(be, {
                    shadow: "hover",
                    style: { width: "240px" }
                  }, {
                    default: t(() => [
                      e(ce, {
                        title: "总收入(元)",
                        value: p.totalRevenue / 100,
                        precision: 2
                      }, null, 8, ["value"])
                    ]),
                    _: 1
                  })
                ]),
                e(fe, {
                  data: p.ranking,
                  stripe: ""
                }, {
                  default: t(() => [
                    e(G, {
                      prop: "productName",
                      label: "DLC名称"
                    }),
                    e(G, {
                      prop: "downloads",
                      label: "下载次数",
                      width: "120"
                    }),
                    e(G, {
                      label: "收入(元)",
                      width: "120"
                    }, {
                      default: t(({ row: c }) => [
                        u(R((c.revenue / 100).toFixed(2)), 1)
                      ]),
                      _: 1
                    })
                  ]),
                  _: 1
                }, 8, ["data"])
              ]),
              _: 1
            })
          ]),
          _: 1
        }, 8, ["modelValue"]),
        e(Ue, {
          modelValue: $.value,
          "onUpdate:modelValue": n[16] || (n[16] = (c) => $.value = c),
          title: b.id ? "编辑DLC" : "新增DLC",
          width: "560px"
        }, {
          footer: t(() => [
            e(Y, {
              onClick: n[15] || (n[15] = (c) => $.value = !1)
            }, {
              default: t(() => [...n[30] || (n[30] = [
                u("取消", -1)
              ])]),
              _: 1
            }),
            e(Y, {
              type: "primary",
              loading: E.value,
              onClick: J
            }, {
              default: t(() => [...n[31] || (n[31] = [
                u("确定", -1)
              ])]),
              _: 1
            }, 8, ["loading"])
          ]),
          default: t(() => [
            e(C, {
              ref_key: "formRef",
              ref: q,
              model: b,
              rules: O,
              "label-width": "110px"
            }, {
              default: t(() => [
                e(m, {
                  label: "游戏",
                  prop: "gameId"
                }, {
                  default: t(() => [
                    e(_, {
                      modelValue: b.gameId,
                      "onUpdate:modelValue": n[6] || (n[6] = (c) => b.gameId = c),
                      placeholder: "请选择游戏",
                      style: { width: "100%" }
                    }, {
                      default: t(() => [
                        (P(!0), Z(ie, null, ve(i.value, (c) => (P(), M(r, {
                          key: c.id,
                          label: c.name,
                          value: c.id
                        }, null, 8, ["label", "value"]))), 128))
                      ]),
                      _: 1
                    }, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(m, {
                  label: "DLC Key",
                  prop: "dlcKey"
                }, {
                  default: t(() => [
                    e(X, {
                      modelValue: b.dlcKey,
                      "onUpdate:modelValue": n[7] || (n[7] = (c) => b.dlcKey = c)
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(m, {
                  label: "DLC名称",
                  prop: "name"
                }, {
                  default: t(() => [
                    e(X, {
                      modelValue: b.name,
                      "onUpdate:modelValue": n[8] || (n[8] = (c) => b.name = c)
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(m, {
                  label: "版本",
                  prop: "version"
                }, {
                  default: t(() => [
                    e(X, {
                      modelValue: b.version,
                      "onUpdate:modelValue": n[9] || (n[9] = (c) => b.version = c)
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(m, { label: "描述" }, {
                  default: t(() => [
                    e(X, {
                      modelValue: b.description,
                      "onUpdate:modelValue": n[10] || (n[10] = (c) => b.description = c),
                      type: "textarea",
                      rows: 3
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(m, { label: "是否免费" }, {
                  default: t(() => [
                    e(ye, {
                      modelValue: b.isFree,
                      "onUpdate:modelValue": n[11] || (n[11] = (c) => b.isFree = c)
                    }, {
                      default: t(() => [
                        e(se, { value: 1 }, {
                          default: t(() => [...n[23] || (n[23] = [
                            u("免费", -1)
                          ])]),
                          _: 1
                        }),
                        e(se, { value: 2 }, {
                          default: t(() => [...n[24] || (n[24] = [
                            u("付费", -1)
                          ])]),
                          _: 1
                        })
                      ]),
                      _: 1
                    }, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(m, {
                  label: "价格(元)",
                  prop: "price"
                }, {
                  default: t(() => [
                    e(xe, {
                      modelValue: b.price,
                      "onUpdate:modelValue": n[12] || (n[12] = (c) => b.price = c),
                      min: 0,
                      precision: 2,
                      disabled: b.isFree === 1,
                      style: { width: "100%" }
                    }, null, 8, ["modelValue", "disabled"])
                  ]),
                  _: 1
                }),
                e(m, { label: "状态" }, {
                  default: t(() => [
                    e(ye, {
                      modelValue: b.status,
                      "onUpdate:modelValue": n[13] || (n[13] = (c) => b.status = c)
                    }, {
                      default: t(() => [
                        e(se, { value: 1 }, {
                          default: t(() => [...n[25] || (n[25] = [
                            u("上架", -1)
                          ])]),
                          _: 1
                        }),
                        e(se, { value: 2 }, {
                          default: t(() => [...n[26] || (n[26] = [
                            u("下架", -1)
                          ])]),
                          _: 1
                        })
                      ]),
                      _: 1
                    }, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(m, { label: "最低游戏版本" }, {
                  default: t(() => [
                    e(X, {
                      modelValue: b.minGameVersion,
                      "onUpdate:modelValue": n[14] || (n[14] = (c) => b.minGameVersion = c)
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                b.id ? (P(), M(m, {
                  key: 0,
                  label: "PCK文件"
                }, {
                  default: t(() => [
                    e(Se, {
                      "auto-upload": !1,
                      limit: 1,
                      accept: ".pck",
                      "on-change": B,
                      "file-list": w.value
                    }, {
                      tip: t(() => [...n[28] || (n[28] = [
                        j("div", { class: "el-upload__tip" }, "仅支持 .pck 文件，大小不超过 500MB", -1)
                      ])]),
                      default: t(() => [
                        e(Y, { type: "primary" }, {
                          default: t(() => [...n[27] || (n[27] = [
                            u("选择文件", -1)
                          ])]),
                          _: 1
                        })
                      ]),
                      _: 1
                    }, 8, ["file-list"]),
                    K.value ? (P(), M(Y, {
                      key: 0,
                      type: "success",
                      loading: k.value,
                      onClick: A,
                      style: { "margin-top": "8px" }
                    }, {
                      default: t(() => [...n[29] || (n[29] = [
                        u("上传", -1)
                      ])]),
                      _: 1
                    }, 8, ["loading"])) : Q("", !0)
                  ]),
                  _: 1
                })) : Q("", !0)
              ]),
              _: 1
            }, 8, ["model"])
          ]),
          _: 1
        }, 8, ["modelValue", "title"])
      ]);
    };
  }
}), je = /* @__PURE__ */ oe(Ae, [["__scopeId", "data-v-80d0f5b7"]]), Ee = { class: "plugin-page" }, ue = "/api/v1/plugin/game/player", qe = /* @__PURE__ */ te({
  __name: "PlayerList",
  setup(h) {
    const L = I(!1), T = I(!1), E = I([]), k = I(0), $ = I(!1), q = I(!1), y = I(), g = I(0), i = H({ gameId: "", uid: "", nickname: "", status: "", pageIndex: 1, pageSize: 10 }), V = H({}), b = H({ banReason: "" }), O = { banReason: [{ required: !0, message: "请输入封禁原因", trigger: "blur" }] };
    function K() {
      try {
        const d = document.cookie.match(/authorized-token=([^;]+)/);
        if (d)
          return JSON.parse(decodeURIComponent(d[1]))?.accessToken || "";
        const o = localStorage.getItem("responsive-user-info");
        if (o)
          return JSON.parse(o)?.accessToken || "";
      } catch {
      }
      return "";
    }
    async function w(d, o, D) {
      const U = {
        Authorization: `Bearer ${K()}`
      }, J = { method: d, headers: U };
      return D && (U["Content-Type"] = "application/json", J.body = JSON.stringify(D)), (await fetch(o, J)).json();
    }
    async function p() {
      L.value = !0;
      try {
        const d = new URLSearchParams();
        d.set("pageIndex", String(i.pageIndex)), d.set("pageSize", String(i.pageSize)), i.gameId && d.set("gameId", i.gameId), i.uid && d.set("uid", i.uid), i.nickname && d.set("nickname", i.nickname), i.status && d.set("status", String(i.status));
        const o = await w("GET", `${ue}?${d.toString()}`);
        o.code === 200 && (E.value = o.data?.list || o.data || [], k.value = o.data?.count || o.count || 0);
      } finally {
        L.value = !1;
      }
    }
    function F() {
      i.pageIndex = 1, p();
    }
    function N() {
      Object.assign(i, { gameId: "", uid: "", nickname: "", status: "", pageIndex: 1 }), p();
    }
    async function x(d) {
      try {
        const o = await w("GET", `${ue}/${d.id}`);
        o.code === 200 ? Object.assign(V, o.data || d) : Object.assign(V, d);
      } catch {
        Object.assign(V, d);
      }
      $.value = !0;
    }
    function a(d) {
      g.value = d.id, b.banReason = "", q.value = !0;
    }
    async function l() {
      await y.value?.validate(), T.value = !0;
      try {
        const d = await w("PUT", `${ue}/${g.value}/ban`, { status: 2, banReason: b.banReason });
        d.code === 200 ? (z.success("封禁成功"), q.value = !1, p()) : z.error(d.msg || "操作失败");
      } finally {
        T.value = !1;
      }
    }
    async function v(d) {
      const o = await w("PUT", `${ue}/${d.id}/ban`, { status: 1, banReason: "" });
      o.code === 200 ? (z.success("解封成功"), p()) : z.error(o.msg || "操作失败");
    }
    return le(p), (d, o) => {
      const D = s("el-input"), U = s("el-form-item"), J = s("el-option"), W = s("el-select"), S = s("el-button"), B = s("el-form"), A = s("el-table-column"), f = s("el-tag"), n = s("el-table"), r = s("el-pagination"), _ = s("el-descriptions-item"), m = s("el-descriptions"), X = s("el-dialog"), Y = ae("loading");
      return P(), Z("div", Ee, [
        o[19] || (o[19] = j("div", { class: "page-header" }, [
          j("h3", null, "玩家管理"),
          j("div")
        ], -1)),
        e(B, {
          inline: !0,
          model: i,
          style: { "margin-bottom": "16px" }
        }, {
          default: t(() => [
            e(U, { label: "游戏ID" }, {
              default: t(() => [
                e(D, {
                  modelValue: i.gameId,
                  "onUpdate:modelValue": o[0] || (o[0] = (C) => i.gameId = C),
                  placeholder: "游戏ID",
                  clearable: ""
                }, null, 8, ["modelValue"])
              ]),
              _: 1
            }),
            e(U, { label: "UID" }, {
              default: t(() => [
                e(D, {
                  modelValue: i.uid,
                  "onUpdate:modelValue": o[1] || (o[1] = (C) => i.uid = C),
                  placeholder: "UID",
                  clearable: ""
                }, null, 8, ["modelValue"])
              ]),
              _: 1
            }),
            e(U, { label: "昵称" }, {
              default: t(() => [
                e(D, {
                  modelValue: i.nickname,
                  "onUpdate:modelValue": o[2] || (o[2] = (C) => i.nickname = C),
                  placeholder: "昵称",
                  clearable: ""
                }, null, 8, ["modelValue"])
              ]),
              _: 1
            }),
            e(U, { label: "状态" }, {
              default: t(() => [
                e(W, {
                  modelValue: i.status,
                  "onUpdate:modelValue": o[3] || (o[3] = (C) => i.status = C),
                  placeholder: "全部",
                  clearable: "",
                  style: { width: "100px" }
                }, {
                  default: t(() => [
                    e(J, {
                      label: "正常",
                      value: 1
                    }),
                    e(J, {
                      label: "封禁",
                      value: 2
                    })
                  ]),
                  _: 1
                }, 8, ["modelValue"])
              ]),
              _: 1
            }),
            e(U, null, {
              default: t(() => [
                e(S, {
                  type: "primary",
                  onClick: F
                }, {
                  default: t(() => [...o[11] || (o[11] = [
                    u("查询", -1)
                  ])]),
                  _: 1
                }),
                e(S, { onClick: N }, {
                  default: t(() => [...o[12] || (o[12] = [
                    u("重置", -1)
                  ])]),
                  _: 1
                })
              ]),
              _: 1
            })
          ]),
          _: 1
        }, 8, ["model"]),
        ne((P(), M(n, {
          data: E.value,
          stripe: ""
        }, {
          default: t(() => [
            e(A, {
              prop: "id",
              label: "ID",
              width: "60"
            }),
            e(A, {
              prop: "gameId",
              label: "游戏ID",
              width: "80"
            }),
            e(A, {
              prop: "uid",
              label: "UID",
              width: "120"
            }),
            e(A, {
              prop: "nickname",
              label: "昵称"
            }),
            e(A, {
              prop: "email",
              label: "邮箱"
            }),
            e(A, {
              prop: "platform",
              label: "平台",
              width: "80"
            }),
            e(A, {
              label: "状态",
              width: "80"
            }, {
              default: t(({ row: C }) => [
                e(f, {
                  type: C.status === 1 ? "success" : "danger"
                }, {
                  default: t(() => [
                    u(R(C.status === 1 ? "正常" : "封禁"), 1)
                  ]),
                  _: 2
                }, 1032, ["type"])
              ]),
              _: 1
            }),
            e(A, {
              prop: "lastLoginAt",
              label: "最后登录时间",
              width: "160"
            }),
            e(A, {
              prop: "lastLoginIp",
              label: "最后登录IP",
              width: "130"
            }),
            e(A, {
              label: "操作",
              width: "180",
              fixed: "right"
            }, {
              default: t(({ row: C }) => [
                e(S, {
                  link: "",
                  type: "primary",
                  onClick: (G) => x(C)
                }, {
                  default: t(() => [...o[13] || (o[13] = [
                    u("详情", -1)
                  ])]),
                  _: 1
                }, 8, ["onClick"]),
                C.status === 1 ? (P(), M(S, {
                  key: 0,
                  link: "",
                  type: "danger",
                  onClick: (G) => a(C)
                }, {
                  default: t(() => [...o[14] || (o[14] = [
                    u("封禁", -1)
                  ])]),
                  _: 1
                }, 8, ["onClick"])) : Q("", !0),
                C.status === 2 ? (P(), M(S, {
                  key: 1,
                  link: "",
                  type: "success",
                  onClick: (G) => v(C)
                }, {
                  default: t(() => [...o[15] || (o[15] = [
                    u("解封", -1)
                  ])]),
                  _: 1
                }, 8, ["onClick"])) : Q("", !0)
              ]),
              _: 1
            })
          ]),
          _: 1
        }, 8, ["data"])), [
          [Y, L.value]
        ]),
        e(r, {
          style: { "margin-top": "16px", display: "flex", "justify-content": "flex-end" },
          "current-page": i.pageIndex,
          "onUpdate:currentPage": o[4] || (o[4] = (C) => i.pageIndex = C),
          "page-size": i.pageSize,
          "onUpdate:pageSize": o[5] || (o[5] = (C) => i.pageSize = C),
          total: k.value,
          layout: "total, sizes, prev, pager, next",
          onChange: p
        }, null, 8, ["current-page", "page-size", "total"]),
        e(X, {
          modelValue: $.value,
          "onUpdate:modelValue": o[7] || (o[7] = (C) => $.value = C),
          title: "玩家详情",
          width: "520px"
        }, {
          footer: t(() => [
            e(S, {
              onClick: o[6] || (o[6] = (C) => $.value = !1)
            }, {
              default: t(() => [...o[16] || (o[16] = [
                u("关闭", -1)
              ])]),
              _: 1
            })
          ]),
          default: t(() => [
            e(m, {
              column: 2,
              border: ""
            }, {
              default: t(() => [
                e(_, { label: "ID" }, {
                  default: t(() => [
                    u(R(V.id), 1)
                  ]),
                  _: 1
                }),
                e(_, { label: "游戏ID" }, {
                  default: t(() => [
                    u(R(V.gameId), 1)
                  ]),
                  _: 1
                }),
                e(_, { label: "UID" }, {
                  default: t(() => [
                    u(R(V.uid), 1)
                  ]),
                  _: 1
                }),
                e(_, { label: "昵称" }, {
                  default: t(() => [
                    u(R(V.nickname), 1)
                  ]),
                  _: 1
                }),
                e(_, {
                  label: "邮箱",
                  span: 2
                }, {
                  default: t(() => [
                    u(R(V.email), 1)
                  ]),
                  _: 1
                }),
                e(_, { label: "平台" }, {
                  default: t(() => [
                    u(R(V.platform), 1)
                  ]),
                  _: 1
                }),
                e(_, { label: "平台ID" }, {
                  default: t(() => [
                    u(R(V.platformId), 1)
                  ]),
                  _: 1
                }),
                e(_, {
                  label: "设备ID",
                  span: 2
                }, {
                  default: t(() => [
                    u(R(V.deviceId), 1)
                  ]),
                  _: 1
                }),
                e(_, { label: "状态" }, {
                  default: t(() => [
                    e(f, {
                      type: V.status === 1 ? "success" : "danger"
                    }, {
                      default: t(() => [
                        u(R(V.status === 1 ? "正常" : "封禁"), 1)
                      ]),
                      _: 1
                    }, 8, ["type"])
                  ]),
                  _: 1
                }),
                e(_, { label: "封禁原因" }, {
                  default: t(() => [
                    u(R(V.banReason || "-"), 1)
                  ]),
                  _: 1
                }),
                e(_, { label: "最后登录时间" }, {
                  default: t(() => [
                    u(R(V.lastLoginAt), 1)
                  ]),
                  _: 1
                }),
                e(_, { label: "最后登录IP" }, {
                  default: t(() => [
                    u(R(V.lastLoginIp), 1)
                  ]),
                  _: 1
                })
              ]),
              _: 1
            })
          ]),
          _: 1
        }, 8, ["modelValue"]),
        e(X, {
          modelValue: q.value,
          "onUpdate:modelValue": o[10] || (o[10] = (C) => q.value = C),
          title: "封禁玩家",
          width: "400px"
        }, {
          footer: t(() => [
            e(S, {
              onClick: o[9] || (o[9] = (C) => q.value = !1)
            }, {
              default: t(() => [...o[17] || (o[17] = [
                u("取消", -1)
              ])]),
              _: 1
            }),
            e(S, {
              type: "danger",
              loading: T.value,
              onClick: l
            }, {
              default: t(() => [...o[18] || (o[18] = [
                u("确认封禁", -1)
              ])]),
              _: 1
            }, 8, ["loading"])
          ]),
          default: t(() => [
            e(B, {
              ref_key: "banFormRef",
              ref: y,
              model: b,
              rules: O,
              "label-width": "80px"
            }, {
              default: t(() => [
                e(U, {
                  label: "封禁原因",
                  prop: "banReason"
                }, {
                  default: t(() => [
                    e(D, {
                      modelValue: b.banReason,
                      "onUpdate:modelValue": o[8] || (o[8] = (C) => b.banReason = C),
                      type: "textarea",
                      rows: 3,
                      placeholder: "请输入封禁原因"
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                })
              ]),
              _: 1
            }, 8, ["model"])
          ]),
          _: 1
        }, 8, ["modelValue"])
      ]);
    };
  }
}), Fe = /* @__PURE__ */ oe(qe, [["__scopeId", "data-v-cfa4dda7"]]), Be = { class: "plugin-page" }, Ge = { class: "page-header" }, Ve = "/api/v1/plugin/game/order", Me = /* @__PURE__ */ te({
  __name: "OrderList",
  setup(h) {
    const L = I(!1), T = I([]), E = I(0), k = H({ gameId: "", orderNo: "", status: "", pageIndex: 1, pageSize: 10 }), $ = (w) => ({ pending: "warning", paid: "success", refunded: "info", failed: "danger" })[w] || "", q = (w) => ({ pending: "待支付", paid: "已支付", refunded: "已退款", failed: "失败" })[w] || w;
    function y() {
      try {
        const w = document.cookie.match(/authorized-token=([^;]+)/);
        if (w)
          return JSON.parse(decodeURIComponent(w[1]))?.accessToken || "";
        const p = localStorage.getItem("responsive-user-info");
        if (p)
          return JSON.parse(p)?.accessToken || "";
      } catch {
      }
      return "";
    }
    async function g(w, p, F) {
      const N = {
        Authorization: `Bearer ${y()}`
      };
      return (await fetch(p, { method: w, headers: N })).json();
    }
    async function i() {
      L.value = !0;
      try {
        const w = new URLSearchParams();
        w.set("pageIndex", String(k.pageIndex)), w.set("pageSize", String(k.pageSize)), k.gameId && w.set("gameId", k.gameId), k.orderNo && w.set("orderNo", k.orderNo), k.status && w.set("status", k.status);
        const p = await g("GET", `${Ve}?${w.toString()}`);
        p.code === 200 && (T.value = p.data?.list || p.data || [], E.value = p.data?.count || p.count || 0);
      } finally {
        L.value = !1;
      }
    }
    function V() {
      k.pageIndex = 1, i();
    }
    function b() {
      Object.assign(k, { gameId: "", orderNo: "", status: "", pageIndex: 1 }), i();
    }
    async function O(w) {
      const p = await g("POST", `${Ve}/${w.id}/refund`);
      p.code === 200 ? (z.success("退款成功"), i()) : z.error(p.msg || "退款失败");
    }
    function K() {
      const w = ["ID", "订单号", "游戏ID", "玩家ID", "商品名称", "金额(元)", "货币", "渠道", "状态", "支付时间"], p = T.value.map((d) => [
        d.id,
        d.orderNo,
        d.gameId,
        d.playerId,
        `"${(d.productName || "").replace(/"/g, '""')}"`,
        (d.amount / 100).toFixed(2),
        d.currency,
        d.channel,
        q(d.status),
        d.paidAt || ""
      ]), F = [w.join(","), ...p.map((d) => d.join(","))].join(`
`), N = new Blob(["\uFEFF" + F], { type: "text/csv;charset=utf-8;" }), x = URL.createObjectURL(N), a = document.createElement("a"), l = /* @__PURE__ */ new Date(), v = `${l.getFullYear()}${String(l.getMonth() + 1).padStart(2, "0")}${String(l.getDate()).padStart(2, "0")}`;
      a.href = x, a.download = `orders_${v}.csv`, a.click(), URL.revokeObjectURL(x);
    }
    return le(i), (w, p) => {
      const F = s("el-button"), N = s("el-input"), x = s("el-form-item"), a = s("el-option"), l = s("el-select"), v = s("el-form"), d = s("el-table-column"), o = s("el-tag"), D = s("el-popconfirm"), U = s("el-table"), J = s("el-pagination"), W = ae("loading");
      return P(), Z("div", Be, [
        j("div", Ge, [
          p[6] || (p[6] = j("h3", null, "订单管理", -1)),
          j("div", null, [
            e(F, {
              type: "success",
              onClick: K
            }, {
              default: t(() => [...p[5] || (p[5] = [
                u("导出 CSV", -1)
              ])]),
              _: 1
            })
          ])
        ]),
        e(v, {
          inline: !0,
          model: k,
          style: { "margin-bottom": "16px" }
        }, {
          default: t(() => [
            e(x, { label: "游戏ID" }, {
              default: t(() => [
                e(N, {
                  modelValue: k.gameId,
                  "onUpdate:modelValue": p[0] || (p[0] = (S) => k.gameId = S),
                  placeholder: "请输入游戏ID",
                  clearable: ""
                }, null, 8, ["modelValue"])
              ]),
              _: 1
            }),
            e(x, { label: "订单号" }, {
              default: t(() => [
                e(N, {
                  modelValue: k.orderNo,
                  "onUpdate:modelValue": p[1] || (p[1] = (S) => k.orderNo = S),
                  placeholder: "请输入订单号",
                  clearable: ""
                }, null, 8, ["modelValue"])
              ]),
              _: 1
            }),
            e(x, { label: "状态" }, {
              default: t(() => [
                e(l, {
                  modelValue: k.status,
                  "onUpdate:modelValue": p[2] || (p[2] = (S) => k.status = S),
                  placeholder: "全部",
                  clearable: "",
                  style: { width: "120px" }
                }, {
                  default: t(() => [
                    e(a, {
                      label: "待支付",
                      value: "pending"
                    }),
                    e(a, {
                      label: "已支付",
                      value: "paid"
                    }),
                    e(a, {
                      label: "已退款",
                      value: "refunded"
                    }),
                    e(a, {
                      label: "失败",
                      value: "failed"
                    })
                  ]),
                  _: 1
                }, 8, ["modelValue"])
              ]),
              _: 1
            }),
            e(x, null, {
              default: t(() => [
                e(F, {
                  type: "primary",
                  onClick: V
                }, {
                  default: t(() => [...p[7] || (p[7] = [
                    u("查询", -1)
                  ])]),
                  _: 1
                }),
                e(F, { onClick: b }, {
                  default: t(() => [...p[8] || (p[8] = [
                    u("重置", -1)
                  ])]),
                  _: 1
                })
              ]),
              _: 1
            })
          ]),
          _: 1
        }, 8, ["model"]),
        ne((P(), M(U, {
          data: T.value,
          stripe: ""
        }, {
          default: t(() => [
            e(d, {
              prop: "id",
              label: "ID",
              width: "60"
            }),
            e(d, {
              prop: "orderNo",
              label: "订单号",
              "min-width": "180"
            }),
            e(d, {
              prop: "gameId",
              label: "游戏ID",
              width: "80"
            }),
            e(d, {
              prop: "playerId",
              label: "玩家ID",
              width: "80"
            }),
            e(d, {
              prop: "productName",
              label: "商品名称",
              "min-width": "120"
            }),
            e(d, {
              label: "金额",
              width: "120"
            }, {
              default: t(({ row: S }) => [
                u(R((S.amount / 100).toFixed(2)) + " " + R(S.currency), 1)
              ]),
              _: 1
            }),
            e(d, {
              prop: "channel",
              label: "支付渠道",
              width: "100"
            }),
            e(d, {
              label: "状态",
              width: "90"
            }, {
              default: t(({ row: S }) => [
                e(o, {
                  type: $(S.status)
                }, {
                  default: t(() => [
                    u(R(q(S.status)), 1)
                  ]),
                  _: 2
                }, 1032, ["type"])
              ]),
              _: 1
            }),
            e(d, {
              prop: "paidAt",
              label: "支付时间",
              width: "160"
            }),
            e(d, {
              label: "操作",
              width: "100",
              fixed: "right"
            }, {
              default: t(({ row: S }) => [
                S.status === "paid" ? (P(), M(D, {
                  key: 0,
                  title: "确认退款？",
                  onConfirm: (B) => O(S)
                }, {
                  reference: t(() => [
                    e(F, {
                      link: "",
                      type: "danger"
                    }, {
                      default: t(() => [...p[9] || (p[9] = [
                        u("退款", -1)
                      ])]),
                      _: 1
                    })
                  ]),
                  _: 1
                }, 8, ["onConfirm"])) : Q("", !0)
              ]),
              _: 1
            })
          ]),
          _: 1
        }, 8, ["data"])), [
          [W, L.value]
        ]),
        e(J, {
          style: { "margin-top": "16px", display: "flex", "justify-content": "flex-end" },
          "current-page": k.pageIndex,
          "onUpdate:currentPage": p[3] || (p[3] = (S) => k.pageIndex = S),
          "page-size": k.pageSize,
          "onUpdate:pageSize": p[4] || (p[4] = (S) => k.pageSize = S),
          total: E.value,
          layout: "total, sizes, prev, pager, next",
          onChange: i
        }, null, 8, ["current-page", "page-size", "total"])
      ]);
    };
  }
}), Je = /* @__PURE__ */ oe(Me, [["__scopeId", "data-v-4d31db61"]]), He = { class: "plugin-page" }, We = { class: "page-header" }, pe = "/api/v1/plugin/game/payment", Xe = /* @__PURE__ */ te({
  __name: "PaymentConfig",
  setup(h) {
    const L = I(!1), T = I(!1), E = I([]), k = I(0), $ = I(!1), q = I(), y = H({ gameId: "", pageIndex: 1, pageSize: 10 }), g = () => ({
      id: null,
      gameId: "",
      channel: "stripe",
      channelName: "",
      enabled: 1,
      env: "sandbox",
      stripePublishableKey: "",
      stripeSecretKey: "",
      stripeWebhookSecret: "",
      alipayAppId: "",
      alipayPrivateKey: "",
      alipayPublicKey: "",
      alipayNotifyUrl: "",
      wechatAppId: "",
      wechatMchId: "",
      wechatApiKey: "",
      wechatNotifyUrl: ""
    }), i = H(g()), V = {
      gameId: [{ required: !0, message: "请输入游戏ID", trigger: "blur" }],
      channel: [{ required: !0, message: "请选择渠道", trigger: "change" }],
      channelName: [{ required: !0, message: "请输入渠道名称", trigger: "blur" }],
      enabled: [{ required: !0, message: "请选择启用状态", trigger: "change" }],
      env: [{ required: !0, message: "请选择环境", trigger: "change" }]
    };
    function b() {
      try {
        const a = document.cookie.match(/authorized-token=([^;]+)/);
        if (a)
          return JSON.parse(decodeURIComponent(a[1]))?.accessToken || "";
        const l = localStorage.getItem("responsive-user-info");
        if (l)
          return JSON.parse(l)?.accessToken || "";
      } catch {
      }
      return "";
    }
    async function O(a, l, v) {
      const d = {
        Authorization: `Bearer ${b()}`
      }, o = { method: a, headers: d };
      return v && (d["Content-Type"] = "application/json", o.body = JSON.stringify(v)), (await fetch(l, o)).json();
    }
    async function K() {
      L.value = !0;
      try {
        const a = new URLSearchParams();
        a.set("pageIndex", String(y.pageIndex)), a.set("pageSize", String(y.pageSize)), y.gameId && a.set("gameId", y.gameId);
        const l = await O("GET", `${pe}?${a.toString()}`);
        l.code === 200 && (E.value = l.data?.list || l.data || [], k.value = l.data?.count || l.count || 0);
      } finally {
        L.value = !1;
      }
    }
    function w() {
      y.pageIndex = 1, K();
    }
    function p() {
      Object.assign(y, { gameId: "", pageIndex: 1 }), K();
    }
    function F() {
      Object.assign(i, {
        stripePublishableKey: "",
        stripeSecretKey: "",
        stripeWebhookSecret: "",
        alipayAppId: "",
        alipayPrivateKey: "",
        alipayPublicKey: "",
        alipayNotifyUrl: "",
        wechatAppId: "",
        wechatMchId: "",
        wechatApiKey: "",
        wechatNotifyUrl: ""
      });
    }
    function N(a) {
      Object.assign(i, g()), a && Object.assign(i, a), $.value = !0;
    }
    async function x() {
      await q.value?.validate(), T.value = !0;
      try {
        const a = i.id ? await O("PUT", `${pe}/${i.id}`, i) : await O("POST", pe, i);
        a.code === 200 ? (z.success("操作成功"), $.value = !1, K()) : z.error(a.msg || "操作失败");
      } finally {
        T.value = !1;
      }
    }
    return le(K), (a, l) => {
      const v = s("el-button"), d = s("el-input"), o = s("el-form-item"), D = s("el-form"), U = s("el-table-column"), J = s("el-tag"), W = s("el-table"), S = s("el-pagination"), B = s("el-option"), A = s("el-select"), f = s("el-dialog"), n = ae("loading");
      return P(), Z("div", He, [
        j("div", We, [
          l[23] || (l[23] = j("h3", null, "支付配置", -1)),
          j("div", null, [
            e(v, {
              type: "primary",
              onClick: l[0] || (l[0] = (r) => N())
            }, {
              default: t(() => [...l[22] || (l[22] = [
                u("新增配置", -1)
              ])]),
              _: 1
            })
          ])
        ]),
        e(D, {
          inline: !0,
          model: y,
          style: { "margin-bottom": "16px" }
        }, {
          default: t(() => [
            e(o, { label: "游戏ID" }, {
              default: t(() => [
                e(d, {
                  modelValue: y.gameId,
                  "onUpdate:modelValue": l[1] || (l[1] = (r) => y.gameId = r),
                  placeholder: "请输入游戏ID",
                  clearable: ""
                }, null, 8, ["modelValue"])
              ]),
              _: 1
            }),
            e(o, null, {
              default: t(() => [
                e(v, {
                  type: "primary",
                  onClick: w
                }, {
                  default: t(() => [...l[24] || (l[24] = [
                    u("查询", -1)
                  ])]),
                  _: 1
                }),
                e(v, { onClick: p }, {
                  default: t(() => [...l[25] || (l[25] = [
                    u("重置", -1)
                  ])]),
                  _: 1
                })
              ]),
              _: 1
            })
          ]),
          _: 1
        }, 8, ["model"]),
        ne((P(), M(W, {
          data: E.value,
          stripe: ""
        }, {
          default: t(() => [
            e(U, {
              prop: "id",
              label: "ID",
              width: "60"
            }),
            e(U, {
              prop: "gameId",
              label: "游戏ID",
              width: "80"
            }),
            e(U, {
              prop: "channel",
              label: "渠道标识",
              width: "100"
            }),
            e(U, {
              prop: "channelName",
              label: "渠道名称"
            }),
            e(U, {
              prop: "env",
              label: "环境",
              width: "100"
            }),
            e(U, {
              label: "启用状态",
              width: "90"
            }, {
              default: t(({ row: r }) => [
                e(J, {
                  type: r.enabled === 1 ? "success" : "danger"
                }, {
                  default: t(() => [
                    u(R(r.enabled === 1 ? "启用" : "禁用"), 1)
                  ]),
                  _: 2
                }, 1032, ["type"])
              ]),
              _: 1
            }),
            e(U, {
              label: "操作",
              width: "100",
              fixed: "right"
            }, {
              default: t(({ row: r }) => [
                e(v, {
                  link: "",
                  type: "primary",
                  onClick: (_) => N(r)
                }, {
                  default: t(() => [...l[26] || (l[26] = [
                    u("编辑", -1)
                  ])]),
                  _: 1
                }, 8, ["onClick"])
              ]),
              _: 1
            })
          ]),
          _: 1
        }, 8, ["data"])), [
          [n, L.value]
        ]),
        e(S, {
          style: { "margin-top": "16px", display: "flex", "justify-content": "flex-end" },
          "current-page": y.pageIndex,
          "onUpdate:currentPage": l[2] || (l[2] = (r) => y.pageIndex = r),
          "page-size": y.pageSize,
          "onUpdate:pageSize": l[3] || (l[3] = (r) => y.pageSize = r),
          total: k.value,
          layout: "total, sizes, prev, pager, next",
          onChange: K
        }, null, 8, ["current-page", "page-size", "total"]),
        e(f, {
          modelValue: $.value,
          "onUpdate:modelValue": l[21] || (l[21] = (r) => $.value = r),
          title: i.id ? "编辑支付配置" : "新增支付配置",
          width: "560px"
        }, {
          footer: t(() => [
            e(v, {
              onClick: l[20] || (l[20] = (r) => $.value = !1)
            }, {
              default: t(() => [...l[27] || (l[27] = [
                u("取消", -1)
              ])]),
              _: 1
            }),
            e(v, {
              type: "primary",
              loading: T.value,
              onClick: x
            }, {
              default: t(() => [...l[28] || (l[28] = [
                u("确定", -1)
              ])]),
              _: 1
            }, 8, ["loading"])
          ]),
          default: t(() => [
            e(D, {
              ref_key: "formRef",
              ref: q,
              model: i,
              rules: V,
              "label-width": "130px"
            }, {
              default: t(() => [
                e(o, {
                  label: "游戏ID",
                  prop: "gameId"
                }, {
                  default: t(() => [
                    e(d, {
                      modelValue: i.gameId,
                      "onUpdate:modelValue": l[4] || (l[4] = (r) => i.gameId = r)
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(o, {
                  label: "渠道",
                  prop: "channel"
                }, {
                  default: t(() => [
                    e(A, {
                      modelValue: i.channel,
                      "onUpdate:modelValue": l[5] || (l[5] = (r) => i.channel = r),
                      style: { width: "100%" },
                      onChange: F
                    }, {
                      default: t(() => [
                        e(B, {
                          label: "Stripe",
                          value: "stripe"
                        }),
                        e(B, {
                          label: "支付宝",
                          value: "alipay"
                        }),
                        e(B, {
                          label: "微信支付",
                          value: "wechat"
                        })
                      ]),
                      _: 1
                    }, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(o, {
                  label: "渠道名称",
                  prop: "channelName"
                }, {
                  default: t(() => [
                    e(d, {
                      modelValue: i.channelName,
                      "onUpdate:modelValue": l[6] || (l[6] = (r) => i.channelName = r)
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(o, {
                  label: "启用状态",
                  prop: "enabled"
                }, {
                  default: t(() => [
                    e(A, {
                      modelValue: i.enabled,
                      "onUpdate:modelValue": l[7] || (l[7] = (r) => i.enabled = r),
                      style: { width: "100%" }
                    }, {
                      default: t(() => [
                        e(B, {
                          label: "启用",
                          value: 1
                        }),
                        e(B, {
                          label: "禁用",
                          value: 2
                        })
                      ]),
                      _: 1
                    }, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(o, {
                  label: "环境",
                  prop: "env"
                }, {
                  default: t(() => [
                    e(A, {
                      modelValue: i.env,
                      "onUpdate:modelValue": l[8] || (l[8] = (r) => i.env = r),
                      style: { width: "100%" }
                    }, {
                      default: t(() => [
                        e(B, {
                          label: "沙箱",
                          value: "sandbox"
                        }),
                        e(B, {
                          label: "生产",
                          value: "production"
                        })
                      ]),
                      _: 1
                    }, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                i.channel === "stripe" ? (P(), Z(ie, { key: 0 }, [
                  e(o, { label: "Publishable Key" }, {
                    default: t(() => [
                      e(d, {
                        modelValue: i.stripePublishableKey,
                        "onUpdate:modelValue": l[9] || (l[9] = (r) => i.stripePublishableKey = r)
                      }, null, 8, ["modelValue"])
                    ]),
                    _: 1
                  }),
                  e(o, { label: "Secret Key" }, {
                    default: t(() => [
                      e(d, {
                        modelValue: i.stripeSecretKey,
                        "onUpdate:modelValue": l[10] || (l[10] = (r) => i.stripeSecretKey = r),
                        "show-password": ""
                      }, null, 8, ["modelValue"])
                    ]),
                    _: 1
                  }),
                  e(o, { label: "Webhook Secret" }, {
                    default: t(() => [
                      e(d, {
                        modelValue: i.stripeWebhookSecret,
                        "onUpdate:modelValue": l[11] || (l[11] = (r) => i.stripeWebhookSecret = r),
                        "show-password": ""
                      }, null, 8, ["modelValue"])
                    ]),
                    _: 1
                  })
                ], 64)) : Q("", !0),
                i.channel === "alipay" ? (P(), Z(ie, { key: 1 }, [
                  e(o, { label: "App ID" }, {
                    default: t(() => [
                      e(d, {
                        modelValue: i.alipayAppId,
                        "onUpdate:modelValue": l[12] || (l[12] = (r) => i.alipayAppId = r)
                      }, null, 8, ["modelValue"])
                    ]),
                    _: 1
                  }),
                  e(o, { label: "应用私钥" }, {
                    default: t(() => [
                      e(d, {
                        modelValue: i.alipayPrivateKey,
                        "onUpdate:modelValue": l[13] || (l[13] = (r) => i.alipayPrivateKey = r),
                        type: "textarea",
                        rows: 4
                      }, null, 8, ["modelValue"])
                    ]),
                    _: 1
                  }),
                  e(o, { label: "支付宝公钥" }, {
                    default: t(() => [
                      e(d, {
                        modelValue: i.alipayPublicKey,
                        "onUpdate:modelValue": l[14] || (l[14] = (r) => i.alipayPublicKey = r),
                        type: "textarea",
                        rows: 4
                      }, null, 8, ["modelValue"])
                    ]),
                    _: 1
                  }),
                  e(o, { label: "回调地址" }, {
                    default: t(() => [
                      e(d, {
                        modelValue: i.alipayNotifyUrl,
                        "onUpdate:modelValue": l[15] || (l[15] = (r) => i.alipayNotifyUrl = r)
                      }, null, 8, ["modelValue"])
                    ]),
                    _: 1
                  })
                ], 64)) : Q("", !0),
                i.channel === "wechat" ? (P(), Z(ie, { key: 2 }, [
                  e(o, { label: "App ID" }, {
                    default: t(() => [
                      e(d, {
                        modelValue: i.wechatAppId,
                        "onUpdate:modelValue": l[16] || (l[16] = (r) => i.wechatAppId = r)
                      }, null, 8, ["modelValue"])
                    ]),
                    _: 1
                  }),
                  e(o, { label: "商户号" }, {
                    default: t(() => [
                      e(d, {
                        modelValue: i.wechatMchId,
                        "onUpdate:modelValue": l[17] || (l[17] = (r) => i.wechatMchId = r)
                      }, null, 8, ["modelValue"])
                    ]),
                    _: 1
                  }),
                  e(o, { label: "API Key" }, {
                    default: t(() => [
                      e(d, {
                        modelValue: i.wechatApiKey,
                        "onUpdate:modelValue": l[18] || (l[18] = (r) => i.wechatApiKey = r),
                        "show-password": ""
                      }, null, 8, ["modelValue"])
                    ]),
                    _: 1
                  }),
                  e(o, { label: "回调地址" }, {
                    default: t(() => [
                      e(d, {
                        modelValue: i.wechatNotifyUrl,
                        "onUpdate:modelValue": l[19] || (l[19] = (r) => i.wechatNotifyUrl = r)
                      }, null, 8, ["modelValue"])
                    ]),
                    _: 1
                  })
                ], 64)) : Q("", !0)
              ]),
              _: 1
            }, 8, ["model"])
          ]),
          _: 1
        }, 8, ["modelValue", "title"])
      ]);
    };
  }
}), Ye = /* @__PURE__ */ oe(Xe, [["__scopeId", "data-v-897573d1"]]), Ze = { class: "plugin-page" }, Qe = { class: "page-header" }, re = "/api/v1/plugin/game/h5", he = /* @__PURE__ */ te({
  __name: "H5List",
  setup(h) {
    const L = I(!1), T = I(!1), E = I([]), k = I(0), $ = I(!1), q = I(), y = H({ gameId: "", name: "", pageIndex: 1, pageSize: 10 }), g = H({
      id: null,
      gameId: "",
      pageKey: "",
      name: "",
      pageType: "custom",
      useExternal: 1,
      content: "",
      externalUrl: "",
      status: 1,
      remark: ""
    }), i = {
      gameId: [{ required: !0, message: "请输入游戏ID", trigger: "blur" }],
      pageKey: [{ required: !0, message: "请输入页面标识", trigger: "blur" }],
      name: [{ required: !0, message: "请输入名称", trigger: "blur" }],
      pageType: [{ required: !0, message: "请选择页面类型", trigger: "change" }]
    };
    function V() {
      try {
        const x = document.cookie.match(/authorized-token=([^;]+)/);
        if (x)
          return JSON.parse(decodeURIComponent(x[1]))?.accessToken || "";
        const a = localStorage.getItem("responsive-user-info");
        if (a)
          return JSON.parse(a)?.accessToken || "";
      } catch {
      }
      return "";
    }
    async function b(x, a, l) {
      const v = {
        Authorization: `Bearer ${V()}`
      }, d = { method: x, headers: v };
      return l && (v["Content-Type"] = "application/json", d.body = JSON.stringify(l)), (await fetch(a, d)).json();
    }
    async function O() {
      L.value = !0;
      try {
        const x = new URLSearchParams();
        x.set("pageIndex", String(y.pageIndex)), x.set("pageSize", String(y.pageSize)), y.gameId && x.set("gameId", y.gameId), y.name && x.set("name", y.name);
        const a = await b("GET", `${re}?${x.toString()}`);
        a.code === 200 && (E.value = a.data?.list || a.data || [], k.value = a.data?.count || a.count || 0);
      } finally {
        L.value = !1;
      }
    }
    function K() {
      y.pageIndex = 1, O();
    }
    function w() {
      Object.assign(y, { gameId: "", name: "", pageIndex: 1 }), O();
    }
    function p(x) {
      Object.assign(g, {
        id: null,
        gameId: "",
        pageKey: "",
        name: "",
        pageType: "custom",
        useExternal: 1,
        content: "",
        externalUrl: "",
        status: 1,
        remark: ""
      }), x && Object.assign(g, x), $.value = !0;
    }
    async function F() {
      await q.value?.validate(), T.value = !0;
      try {
        const x = g.id ? await b("PUT", `${re}/${g.id}`, g) : await b("POST", re, g);
        x.code === 200 ? (z.success("操作成功"), $.value = !1, O()) : z.error(x.msg || "操作失败");
      } finally {
        T.value = !1;
      }
    }
    async function N(x) {
      const a = await b("DELETE", `${re}/${x.id}`);
      a.code === 200 ? (z.success("删除成功"), O()) : z.error(a.msg || "删除失败");
    }
    return le(O), (x, a) => {
      const l = s("el-button"), v = s("el-input"), d = s("el-form-item"), o = s("el-form"), D = s("el-table-column"), U = s("el-tag"), J = s("el-popconfirm"), W = s("el-table"), S = s("el-pagination"), B = s("el-option"), A = s("el-select"), f = s("el-radio"), n = s("el-radio-group"), r = s("el-dialog"), _ = ae("loading");
      return P(), Z("div", Ze, [
        j("div", Qe, [
          a[17] || (a[17] = j("h3", null, "H5页面管理", -1)),
          j("div", null, [
            e(l, {
              type: "primary",
              onClick: a[0] || (a[0] = (m) => p())
            }, {
              default: t(() => [...a[16] || (a[16] = [
                u("新增H5页面", -1)
              ])]),
              _: 1
            })
          ])
        ]),
        e(o, {
          inline: !0,
          model: y,
          style: { "margin-bottom": "16px" }
        }, {
          default: t(() => [
            e(d, { label: "游戏ID" }, {
              default: t(() => [
                e(v, {
                  modelValue: y.gameId,
                  "onUpdate:modelValue": a[1] || (a[1] = (m) => y.gameId = m),
                  placeholder: "请输入游戏ID",
                  clearable: ""
                }, null, 8, ["modelValue"])
              ]),
              _: 1
            }),
            e(d, { label: "页面名称" }, {
              default: t(() => [
                e(v, {
                  modelValue: y.name,
                  "onUpdate:modelValue": a[2] || (a[2] = (m) => y.name = m),
                  placeholder: "请输入页面名称",
                  clearable: ""
                }, null, 8, ["modelValue"])
              ]),
              _: 1
            }),
            e(d, null, {
              default: t(() => [
                e(l, {
                  type: "primary",
                  onClick: K
                }, {
                  default: t(() => [...a[18] || (a[18] = [
                    u("查询", -1)
                  ])]),
                  _: 1
                }),
                e(l, { onClick: w }, {
                  default: t(() => [...a[19] || (a[19] = [
                    u("重置", -1)
                  ])]),
                  _: 1
                })
              ]),
              _: 1
            })
          ]),
          _: 1
        }, 8, ["model"]),
        ne((P(), M(W, {
          data: E.value,
          stripe: ""
        }, {
          default: t(() => [
            e(D, {
              prop: "id",
              label: "ID",
              width: "60"
            }),
            e(D, {
              prop: "gameId",
              label: "游戏ID",
              width: "80"
            }),
            e(D, {
              prop: "pageKey",
              label: "页面标识"
            }),
            e(D, {
              prop: "name",
              label: "名称"
            }),
            e(D, {
              prop: "pageType",
              label: "页面类型",
              width: "100"
            }),
            e(D, {
              label: "是否外链",
              width: "90"
            }, {
              default: t(({ row: m }) => [
                e(U, {
                  type: m.useExternal === 1 ? "warning" : "info"
                }, {
                  default: t(() => [
                    u(R(m.useExternal === 1 ? "外链" : "内嵌"), 1)
                  ]),
                  _: 2
                }, 1032, ["type"])
              ]),
              _: 1
            }),
            e(D, {
              label: "状态",
              width: "80"
            }, {
              default: t(({ row: m }) => [
                e(U, {
                  type: m.status === 1 ? "success" : "danger"
                }, {
                  default: t(() => [
                    u(R(m.status === 1 ? "启用" : "禁用"), 1)
                  ]),
                  _: 2
                }, 1032, ["type"])
              ]),
              _: 1
            }),
            e(D, {
              label: "操作",
              width: "160",
              fixed: "right"
            }, {
              default: t(({ row: m }) => [
                e(l, {
                  link: "",
                  type: "primary",
                  onClick: (X) => p(m)
                }, {
                  default: t(() => [...a[20] || (a[20] = [
                    u("编辑", -1)
                  ])]),
                  _: 1
                }, 8, ["onClick"]),
                e(J, {
                  title: "确认删除？",
                  onConfirm: (X) => N(m)
                }, {
                  reference: t(() => [
                    e(l, {
                      link: "",
                      type: "danger"
                    }, {
                      default: t(() => [...a[21] || (a[21] = [
                        u("删除", -1)
                      ])]),
                      _: 1
                    })
                  ]),
                  _: 1
                }, 8, ["onConfirm"])
              ]),
              _: 1
            })
          ]),
          _: 1
        }, 8, ["data"])), [
          [_, L.value]
        ]),
        e(S, {
          style: { "margin-top": "16px", display: "flex", "justify-content": "flex-end" },
          "current-page": y.pageIndex,
          "onUpdate:currentPage": a[3] || (a[3] = (m) => y.pageIndex = m),
          "page-size": y.pageSize,
          "onUpdate:pageSize": a[4] || (a[4] = (m) => y.pageSize = m),
          total: k.value,
          layout: "total, sizes, prev, pager, next",
          onChange: O
        }, null, 8, ["current-page", "page-size", "total"]),
        e(r, {
          modelValue: $.value,
          "onUpdate:modelValue": a[15] || (a[15] = (m) => $.value = m),
          title: g.id ? "编辑H5页面" : "新增H5页面",
          width: "560px"
        }, {
          footer: t(() => [
            e(l, {
              onClick: a[14] || (a[14] = (m) => $.value = !1)
            }, {
              default: t(() => [...a[26] || (a[26] = [
                u("取消", -1)
              ])]),
              _: 1
            }),
            e(l, {
              type: "primary",
              loading: T.value,
              onClick: F
            }, {
              default: t(() => [...a[27] || (a[27] = [
                u("确定", -1)
              ])]),
              _: 1
            }, 8, ["loading"])
          ]),
          default: t(() => [
            e(o, {
              ref_key: "formRef",
              ref: q,
              model: g,
              rules: i,
              "label-width": "90px"
            }, {
              default: t(() => [
                e(d, {
                  label: "游戏ID",
                  prop: "gameId"
                }, {
                  default: t(() => [
                    e(v, {
                      modelValue: g.gameId,
                      "onUpdate:modelValue": a[5] || (a[5] = (m) => g.gameId = m),
                      modelModifiers: { number: !0 }
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(d, {
                  label: "页面标识",
                  prop: "pageKey"
                }, {
                  default: t(() => [
                    e(v, {
                      modelValue: g.pageKey,
                      "onUpdate:modelValue": a[6] || (a[6] = (m) => g.pageKey = m)
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(d, {
                  label: "名称",
                  prop: "name"
                }, {
                  default: t(() => [
                    e(v, {
                      modelValue: g.name,
                      "onUpdate:modelValue": a[7] || (a[7] = (m) => g.name = m)
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(d, {
                  label: "页面类型",
                  prop: "pageType"
                }, {
                  default: t(() => [
                    e(A, {
                      modelValue: g.pageType,
                      "onUpdate:modelValue": a[8] || (a[8] = (m) => g.pageType = m),
                      style: { width: "100%" }
                    }, {
                      default: t(() => [
                        e(B, {
                          label: "custom",
                          value: "custom"
                        }),
                        e(B, {
                          label: "商城页",
                          value: "商城页"
                        }),
                        e(B, {
                          label: "活动页",
                          value: "活动页"
                        }),
                        e(B, {
                          label: "公告页",
                          value: "公告页"
                        })
                      ]),
                      _: 1
                    }, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(d, {
                  label: "链接类型",
                  prop: "useExternal"
                }, {
                  default: t(() => [
                    e(n, {
                      modelValue: g.useExternal,
                      "onUpdate:modelValue": a[9] || (a[9] = (m) => g.useExternal = m)
                    }, {
                      default: t(() => [
                        e(f, { value: 1 }, {
                          default: t(() => [...a[22] || (a[22] = [
                            u("外链", -1)
                          ])]),
                          _: 1
                        }),
                        e(f, { value: 2 }, {
                          default: t(() => [...a[23] || (a[23] = [
                            u("内嵌", -1)
                          ])]),
                          _: 1
                        })
                      ]),
                      _: 1
                    }, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                g.useExternal === 1 ? (P(), M(d, {
                  key: 0,
                  label: "外链地址",
                  prop: "externalUrl"
                }, {
                  default: t(() => [
                    e(v, {
                      modelValue: g.externalUrl,
                      "onUpdate:modelValue": a[10] || (a[10] = (m) => g.externalUrl = m),
                      placeholder: "请输入外链URL"
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                })) : Q("", !0),
                g.useExternal === 2 ? (P(), M(d, {
                  key: 1,
                  label: "页面内容",
                  prop: "content"
                }, {
                  default: t(() => [
                    e(v, {
                      modelValue: g.content,
                      "onUpdate:modelValue": a[11] || (a[11] = (m) => g.content = m),
                      type: "textarea",
                      rows: 8,
                      placeholder: "请输入页面内容"
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                })) : Q("", !0),
                e(d, { label: "状态" }, {
                  default: t(() => [
                    e(n, {
                      modelValue: g.status,
                      "onUpdate:modelValue": a[12] || (a[12] = (m) => g.status = m)
                    }, {
                      default: t(() => [
                        e(f, { value: 1 }, {
                          default: t(() => [...a[24] || (a[24] = [
                            u("启用", -1)
                          ])]),
                          _: 1
                        }),
                        e(f, { value: 2 }, {
                          default: t(() => [...a[25] || (a[25] = [
                            u("禁用", -1)
                          ])]),
                          _: 1
                        })
                      ]),
                      _: 1
                    }, 8, ["modelValue"])
                  ]),
                  _: 1
                }),
                e(d, { label: "备注" }, {
                  default: t(() => [
                    e(v, {
                      modelValue: g.remark,
                      "onUpdate:modelValue": a[13] || (a[13] = (m) => g.remark = m),
                      type: "textarea",
                      rows: 2
                    }, null, 8, ["modelValue"])
                  ]),
                  _: 1
                })
              ]),
              _: 1
            }, 8, ["model"])
          ]),
          _: 1
        }, 8, ["modelValue", "title"])
      ]);
    };
  }
}), et = /* @__PURE__ */ oe(he, [["__scopeId", "data-v-7805674b"]]), at = [
  { path: "/plugin/game/list", name: "PluginGameList", component: Te, meta: { title: "游戏列表" } },
  { path: "/plugin/game/dlc", name: "PluginGameDlc", component: je, meta: { title: "DLC管理" } },
  { path: "/plugin/game/player", name: "PluginGamePlayer", component: Fe, meta: { title: "玩家管理" } },
  { path: "/plugin/game/order", name: "PluginGameOrder", component: Je, meta: { title: "订单管理" } },
  { path: "/plugin/game/payment", name: "PluginGamePayment", component: Ye, meta: { title: "支付配置" } },
  { path: "/plugin/game/h5", name: "PluginGameH5", component: et, meta: { title: "H5页面" } }
], nt = [];
export {
  nt as menus,
  at as routes
};
