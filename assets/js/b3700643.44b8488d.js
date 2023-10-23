"use strict";(self.webpackChunkcompass=self.webpackChunkcompass||[]).push([[533],{3905:function(e,t,n){n.d(t,{Zo:function(){return s},kt:function(){return f}});var r=n(7294);function o(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function c(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,r)}return n}function i(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?c(Object(n),!0).forEach((function(t){o(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):c(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function a(e,t){if(null==e)return{};var n,r,o=function(e,t){if(null==e)return{};var n,r,o={},c=Object.keys(e);for(r=0;r<c.length;r++)n=c[r],t.indexOf(n)>=0||(o[n]=e[n]);return o}(e,t);if(Object.getOwnPropertySymbols){var c=Object.getOwnPropertySymbols(e);for(r=0;r<c.length;r++)n=c[r],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(o[n]=e[n])}return o}var l=r.createContext({}),u=function(e){var t=r.useContext(l),n=t;return e&&(n="function"==typeof e?e(t):i(i({},t),e)),n},s=function(e){var t=u(e.components);return r.createElement(l.Provider,{value:t},e.children)},p={inlineCode:"code",wrapper:function(e){var t=e.children;return r.createElement(r.Fragment,{},t)}},m=r.forwardRef((function(e,t){var n=e.components,o=e.mdxType,c=e.originalType,l=e.parentName,s=a(e,["components","mdxType","originalType","parentName"]),m=u(n),f=o,d=m["".concat(l,".").concat(f)]||m[f]||p[f]||c;return n?r.createElement(d,i(i({ref:t},s),{},{components:n})):r.createElement(d,i({ref:t},s))}));function f(e,t){var n=arguments,o=t&&t.mdxType;if("string"==typeof e||o){var c=n.length,i=new Array(c);i[0]=m;var a={};for(var l in t)hasOwnProperty.call(t,l)&&(a[l]=t[l]);a.originalType=e,a.mdxType="string"==typeof e?e:o,i[1]=a;for(var u=2;u<c;u++)i[u]=n[u];return r.createElement.apply(null,i)}return r.createElement.apply(null,n)}m.displayName="MDXCreateElement"},9099:function(e,t,n){n.r(t),n.d(t,{assets:function(){return s},contentTitle:function(){return l},default:function(){return f},frontMatter:function(){return a},metadata:function(){return u},toc:function(){return p}});var r=n(7462),o=n(3366),c=(n(7294),n(3905)),i=["components"],a={},l="Telemetry",u={unversionedId:"guides/telemetry",id:"guides/telemetry",title:"Telemetry",description:"Compass collects basic HTTP metrics (response time, duration, etc) and sends it to New Relic when enabled.",source:"@site/docs/guides/telemetry.md",sourceDirName:"guides",slug:"/guides/telemetry",permalink:"/compass/guides/telemetry",draft:!1,editUrl:"https://github.com/goto/compass/edit/master/docs/docs/guides/telemetry.md",tags:[],version:"current",frontMatter:{}},s={},p=[{value:"New Relic",id:"new-relic",level:2}],m={toc:p};function f(e){var t=e.components,n=(0,o.Z)(e,i);return(0,c.kt)("wrapper",(0,r.Z)({},m,n,{components:t,mdxType:"MDXLayout"}),(0,c.kt)("h1",{id:"telemetry"},"Telemetry"),(0,c.kt)("p",null,"Compass collects basic HTTP metrics (response time, duration, etc) and sends it to ",(0,c.kt)("a",{parentName:"p",href:"https://newrelic.com/"},"New Relic")," when enabled."),(0,c.kt)("h2",{id:"new-relic"},"New Relic"),(0,c.kt)("p",null,"New Relic is not enabled by default. To enable New Relic, you can set these configurations"),(0,c.kt)("pre",null,(0,c.kt)("code",{parentName:"pre"},"NEW_RELIC_LICENSE_KEY=mf9d13c838u252252c43ji47q1u4ynzpDDDDTSPQ\nNEW_RELIC_APP_NAME=compass\n")),(0,c.kt)("p",null,"Empty ",(0,c.kt)("inlineCode",{parentName:"p"},"NEW_RELIC_LICENSE_KEY")," will disable New Relic integration."))}f.isMDXComponent=!0}}]);