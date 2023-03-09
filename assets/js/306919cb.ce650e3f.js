"use strict";(self.webpackChunkcompass=self.webpackChunkcompass||[]).push([[825],{3905:function(e,t,n){n.d(t,{Zo:function(){return d},kt:function(){return s}});var a=n(7294);function l(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function i(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);t&&(a=a.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,a)}return n}function r(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?i(Object(n),!0).forEach((function(t){l(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):i(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function o(e,t){if(null==e)return{};var n,a,l=function(e,t){if(null==e)return{};var n,a,l={},i=Object.keys(e);for(a=0;a<i.length;a++)n=i[a],t.indexOf(n)>=0||(l[n]=e[n]);return l}(e,t);if(Object.getOwnPropertySymbols){var i=Object.getOwnPropertySymbols(e);for(a=0;a<i.length;a++)n=i[a],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(l[n]=e[n])}return l}var p=a.createContext({}),u=function(e){var t=a.useContext(p),n=t;return e&&(n="function"==typeof e?e(t):r(r({},t),e)),n},d=function(e){var t=u(e.components);return a.createElement(p.Provider,{value:t},e.children)},m={inlineCode:"code",wrapper:function(e){var t=e.children;return a.createElement(a.Fragment,{},t)}},c=a.forwardRef((function(e,t){var n=e.components,l=e.mdxType,i=e.originalType,p=e.parentName,d=o(e,["components","mdxType","originalType","parentName"]),c=u(n),s=l,k=c["".concat(p,".").concat(s)]||c[s]||m[s]||i;return n?a.createElement(k,r(r({ref:t},d),{},{components:n})):a.createElement(k,r({ref:t},d))}));function s(e,t){var n=arguments,l=t&&t.mdxType;if("string"==typeof e||l){var i=n.length,r=new Array(i);r[0]=c;var o={};for(var p in t)hasOwnProperty.call(t,p)&&(o[p]=t[p]);o.originalType=e,o.mdxType="string"==typeof e?e:l,r[1]=o;for(var u=2;u<i;u++)r[u]=n[u];return a.createElement.apply(null,r)}return a.createElement.apply(null,n)}c.displayName="MDXCreateElement"},4480:function(e,t,n){n.r(t),n.d(t,{assets:function(){return d},contentTitle:function(){return p},default:function(){return s},frontMatter:function(){return o},metadata:function(){return u},toc:function(){return m}});var a=n(7462),l=n(3366),i=(n(7294),n(3905)),r=["components"],o={},p="Configurations",u={unversionedId:"reference/configuration",id:"reference/configuration",title:"Configurations",description:"This page contains reference for all the application configurations for Compass.",source:"@site/docs/reference/configuration.md",sourceDirName:"reference",slug:"/reference/configuration",permalink:"/compass/reference/configuration",draft:!1,editUrl:"https://github.com/odpf/compass/edit/master/docs/docs/reference/configuration.md",tags:[],version:"current",frontMatter:{},sidebar:"docsSidebar",previous:{title:"CLI",permalink:"/compass/reference/cli"},next:{title:"Contribution Process",permalink:"/compass/contribute/contributing"}},d={},m=[{value:"Table of Contents",id:"table-of-contents",level:2},{value:"Generic",id:"generic",level:2},{value:"<code>LOG_LEVEL</code>",id:"log_level",level:3},{value:"<code>SERVER_HOST</code>",id:"server_host",level:3},{value:"<code>SERVER_PORT</code>",id:"server_port",level:3},{value:"<code>ELASTICSEARCH_BROKERS</code>",id:"elasticsearch_brokers",level:3},{value:"<code>DB_HOST</code>",id:"db_host",level:3},{value:"<code>DB_PORT</code>",id:"db_port",level:3},{value:"<code>DB_NAME</code>",id:"db_name",level:3},{value:"<code>DB_USER</code>",id:"db_user",level:3},{value:"<code>DB_PASSWORD</code>",id:"db_password",level:3},{value:"<code>DB_SSL_MODE</code>",id:"db_ssl_mode",level:3},{value:"<code>IDENTITY_UUID_HEADER</code>",id:"identity_uuid_header",level:3},{value:"<code>IDENTITY_EMAIL_HEADER</code>",id:"identity_email_header",level:3},{value:"<code>IDENTITY_PROVIDER_DEFAULT_NAME</code>",id:"identity_provider_default_name",level:3},{value:"Telemetry",id:"telemetry",level:2},{value:"<code>STATSD_ADDRESS</code>",id:"statsd_address",level:3},{value:"<code>STATSD_PREFIX</code>",id:"statsd_prefix",level:3},{value:"<code>STATSD_ENABLED</code>",id:"statsd_enabled",level:3},{value:"<code>NEW_RELIC_APP_NAME</code>",id:"new_relic_app_name",level:3},{value:"<code>NEW_RELIC_LICENSE_KEY</code>",id:"new_relic_license_key",level:3}],c={toc:m};function s(e){var t=e.components,n=(0,l.Z)(e,r);return(0,i.kt)("wrapper",(0,a.Z)({},c,n,{components:t,mdxType:"MDXLayout"}),(0,i.kt)("h1",{id:"configurations"},"Configurations"),(0,i.kt)("p",null,"This page contains reference for all the application configurations for Compass."),(0,i.kt)("h2",{id:"table-of-contents"},"Table of Contents"),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"/compass/configuration#-generic"},"Generic")),(0,i.kt)("li",{parentName:"ul"},(0,i.kt)("a",{parentName:"li",href:"/compass/configuration#-telemetry"},"Telemetry"))),(0,i.kt)("h2",{id:"generic"},"Generic"),(0,i.kt)("p",null,"Compass's required variables to start using it."),(0,i.kt)("h3",{id:"log_level"},(0,i.kt)("inlineCode",{parentName:"h3"},"LOG_LEVEL")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"error")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"optional")),(0,i.kt)("li",{parentName:"ul"},"Default: ",(0,i.kt)("inlineCode",{parentName:"li"},"info")),(0,i.kt)("li",{parentName:"ul"},"Logging level, can be one of ",(0,i.kt)("inlineCode",{parentName:"li"},"trace"),", ",(0,i.kt)("inlineCode",{parentName:"li"},"debug"),", ",(0,i.kt)("inlineCode",{parentName:"li"},"info"),", ",(0,i.kt)("inlineCode",{parentName:"li"},"warning"),", ",(0,i.kt)("inlineCode",{parentName:"li"},"error"),", ",(0,i.kt)("inlineCode",{parentName:"li"},"fatal"),", ",(0,i.kt)("inlineCode",{parentName:"li"},"panic"),".")),(0,i.kt)("h3",{id:"server_host"},(0,i.kt)("inlineCode",{parentName:"h3"},"SERVER_HOST")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"localhost")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"required")),(0,i.kt)("li",{parentName:"ul"},"Network interface to bind to.")),(0,i.kt)("h3",{id:"server_port"},(0,i.kt)("inlineCode",{parentName:"h3"},"SERVER_PORT")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"8080")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"required")),(0,i.kt)("li",{parentName:"ul"},"Port to listen on.")),(0,i.kt)("h3",{id:"elasticsearch_brokers"},(0,i.kt)("inlineCode",{parentName:"h3"},"ELASTICSEARCH_BROKERS")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"http://localhost:9200,http://localhost:9300")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"required")),(0,i.kt)("li",{parentName:"ul"},"Comma separated list of elasticsearch nodes.")),(0,i.kt)("h3",{id:"db_host"},(0,i.kt)("inlineCode",{parentName:"h3"},"DB_HOST")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"localhost")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"required")),(0,i.kt)("li",{parentName:"ul"},"PostgreSQL DB hostname to connect.")),(0,i.kt)("h3",{id:"db_port"},(0,i.kt)("inlineCode",{parentName:"h3"},"DB_PORT")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"5432")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"required")),(0,i.kt)("li",{parentName:"ul"},"PostgreSQL DB port to connect.")),(0,i.kt)("h3",{id:"db_name"},(0,i.kt)("inlineCode",{parentName:"h3"},"DB_NAME")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"compass")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"required")),(0,i.kt)("li",{parentName:"ul"},"PostgreSQL DB name to connect.")),(0,i.kt)("h3",{id:"db_user"},(0,i.kt)("inlineCode",{parentName:"h3"},"DB_USER")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"postgres")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"required")),(0,i.kt)("li",{parentName:"ul"},"PostgreSQL DB user to connect.")),(0,i.kt)("h3",{id:"db_password"},(0,i.kt)("inlineCode",{parentName:"h3"},"DB_PASSWORD")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"~")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"required")),(0,i.kt)("li",{parentName:"ul"},"PostgreSQL DB user's password to connect.")),(0,i.kt)("h3",{id:"db_ssl_mode"},(0,i.kt)("inlineCode",{parentName:"h3"},"DB_SSL_MODE")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"disable")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"optional")),(0,i.kt)("li",{parentName:"ul"},"PostgreSQL DB SSL mode to connect.")),(0,i.kt)("h3",{id:"identity_uuid_header"},(0,i.kt)("inlineCode",{parentName:"h3"},"IDENTITY_UUID_HEADER")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"Compass-User-UUID")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"required")),(0,i.kt)("li",{parentName:"ul"},"Header key to accept Compass User UUID. See ",(0,i.kt)("a",{parentName:"li",href:"/compass/concepts/user"},"User")," for more information about the usage.")),(0,i.kt)("h3",{id:"identity_email_header"},(0,i.kt)("inlineCode",{parentName:"h3"},"IDENTITY_EMAIL_HEADER")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"Compass-User-Email")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"optional")),(0,i.kt)("li",{parentName:"ul"},"Header key to accept Compass User Email. See ",(0,i.kt)("a",{parentName:"li",href:"/compass/concepts/user"},"User")," for more information about the usage.")),(0,i.kt)("h3",{id:"identity_provider_default_name"},(0,i.kt)("inlineCode",{parentName:"h3"},"IDENTITY_PROVIDER_DEFAULT_NAME")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"shield")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"optional")),(0,i.kt)("li",{parentName:"ul"},"Default value of user provider. See ",(0,i.kt)("a",{parentName:"li",href:"/compass/concepts/user"},"User")," for more information about the usage.")),(0,i.kt)("h2",{id:"telemetry"},"Telemetry"),(0,i.kt)("p",null,"Variables for metrics gathering."),(0,i.kt)("h3",{id:"statsd_address"},(0,i.kt)("inlineCode",{parentName:"h3"},"STATSD_ADDRESS")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"127.0.0.1:8125")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"optional")),(0,i.kt)("li",{parentName:"ul"},"statsd client to send metrics to.")),(0,i.kt)("h3",{id:"statsd_prefix"},(0,i.kt)("inlineCode",{parentName:"h3"},"STATSD_PREFIX")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"discovery")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"optional")),(0,i.kt)("li",{parentName:"ul"},"Default: ",(0,i.kt)("inlineCode",{parentName:"li"},"compassApi")),(0,i.kt)("li",{parentName:"ul"},"Prefix for statsd metrics names.")),(0,i.kt)("h3",{id:"statsd_enabled"},(0,i.kt)("inlineCode",{parentName:"h3"},"STATSD_ENABLED")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"true")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"required")),(0,i.kt)("li",{parentName:"ul"},"Default: ",(0,i.kt)("inlineCode",{parentName:"li"},"false")),(0,i.kt)("li",{parentName:"ul"},"Enable publishing application metrics to statsd.")),(0,i.kt)("h3",{id:"new_relic_app_name"},(0,i.kt)("inlineCode",{parentName:"h3"},"NEW_RELIC_APP_NAME")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"compass-integration")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"optional")),(0,i.kt)("li",{parentName:"ul"},"Default: ",(0,i.kt)("inlineCode",{parentName:"li"},"compass")),(0,i.kt)("li",{parentName:"ul"},"New Relic application name.")),(0,i.kt)("h3",{id:"new_relic_license_key"},(0,i.kt)("inlineCode",{parentName:"h3"},"NEW_RELIC_LICENSE_KEY")),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"Example value: ",(0,i.kt)("inlineCode",{parentName:"li"},"mf9d13c838u252252c43ji47q1u4ynzpDDDDTSPQ")),(0,i.kt)("li",{parentName:"ul"},"Type: ",(0,i.kt)("inlineCode",{parentName:"li"},"optional")),(0,i.kt)("li",{parentName:"ul"},"Default: ",(0,i.kt)("inlineCode",{parentName:"li"},'""')),(0,i.kt)("li",{parentName:"ul"},"New Relic license key. Empty value would disable newrelic monitoring.")))}s.isMDXComponent=!0}}]);