(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-49342a35"],{5544:function(t,e,a){"use strict";a.r(e);var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"app-container"},[a("el-row",{attrs:{type:"flex",justify:"end"}},[a("el-col",{attrs:{span:2}},[a("el-button",{attrs:{type:"info"},on:{click:t.refresh}},[t._v("刷新")])],1)],1),t._v(" "),a("br"),t._v(" "),a("el-table",{directives:[{name:"loading",rawName:"v-loading",value:t.listLoading,expression:"listLoading"}],attrs:{data:t.workers,"element-loading-text":"Loading",border:"",fit:"","highlight-current-row":""}},[a("el-table-column",{attrs:{align:"center",label:"ID",width:"250"},scopedSlots:t._u([{key:"default",fn:function(e){return[t._v("\n        "+t._s(e.$index)+"\n      ")]}}])}),t._v(" "),a("el-table-column",{attrs:{label:"节点名称",align:"center"},scopedSlots:t._u([{key:"default",fn:function(e){return[t._v("\n        "+t._s(e.row.name)+"\n      ")]}}])})],1)],1)},r=[],s=a("cc08"),o={data:function(){return{workers:[],workersTotal:0,isAdmin:this.$store.getters.user.isAdmin,searchParams:{page_size:200,page:1}}},created:function(){this.fetchData()},methods:{fetchData:function(){var t=this,e=arguments.length>0&&void 0!==arguments[0]?arguments[0]:null;s["a"].list(this.searchParams,function(a){t.workers=a.data,e&&e()})},refresh:function(){var t=this;this.fetchData(function(){t.$message.success("刷新成功")})}}},i=o,c=a("2877"),u=Object(c["a"])(i,n,r,!1,null,null,null);e["default"]=u.exports},cc08:function(t,e,a){"use strict";var n=a("d81d");e["a"]={list:function(t,e){n["a"].get("/worker/list",t,e)},ping:function(t,e){n["a"].get("/worker/ping/".concat(t),{},e)}}},d81d:function(t,e,a){"use strict";a("ac4d"),a("8a81"),a("ac6a");var n=a("bc3a"),r=a.n(n),s=a("5c96"),o=a("a18c"),i="加载失败, 请稍后再试",c=0,u=401,l=801;function d(t,e){t.then(function(t){return p(t,e)}).catch(function(t){return h(t)})}function f(t,e){switch(t){case l:return o["a"].push("/install"),!1;case u:return o["a"].push("/user/login"),!1}return t===c||(s["Message"].error({message:e}),!1)}function p(t,e){f(t.data.code,t.data.message)&&e&&e(t.data.data,t.data.code,t.data.message)}function h(t){s["Message"].error({message:"请求失败 - "+t})}r.a.defaults.baseURL="",r.a.defaults.timeout=1e4,r.a.defaults.responseType="json",r.a.interceptors.response.use(function(t){return t},function(t){return s["Message"].error({message:i}),Promise.reject(t)}),e["a"]={get:function(t,e,a){var n=r.a.get(t,{params:e});d(n,a)},batchGet:function(t,e){var a=[],n=!0,s=!1,o=void 0;try{for(var i,c=t[Symbol.iterator]();!(n=(i=c.next()).done);n=!0){var u=i.value,l={};void 0!==u.params&&(l=u.params),a.push(r.a.get(u.uri,{params:l}))}}catch(d){s=!0,o=d}finally{try{n||null==c.return||c.return()}finally{if(s)throw o}}r.a.all(a).then(r.a.spread(function(){for(var t=[],a=arguments.length,n=new Array(a),r=0;r<a;r++)n[r]=arguments[r];for(var s=0,o=n;s<o.length;s++){var i=o[s];if(!f(i.data.code,i.data.message))return;t.push(i.data.data)}e.apply(void 0,t)})).catch(function(t){return h(t)})},post:function(t,e,a){var n=r.a.post(t,JSON.stringify(e),{headers:{post:{"Content-Type":"application/json"}}});d(n,a)}}}}]);