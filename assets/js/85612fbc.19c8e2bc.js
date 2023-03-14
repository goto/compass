"use strict";(self.webpackChunkcompass=self.webpackChunkcompass||[]).push([[844],{3905:function(e,t,n){n.d(t,{Zo:function(){return o},kt:function(){return m}});var a=n(7294);function i(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function s(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);t&&(a=a.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,a)}return n}function r(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?s(Object(n),!0).forEach((function(t){i(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):s(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function l(e,t){if(null==e)return{};var n,a,i=function(e,t){if(null==e)return{};var n,a,i={},s=Object.keys(e);for(a=0;a<s.length;a++)n=s[a],t.indexOf(n)>=0||(i[n]=e[n]);return i}(e,t);if(Object.getOwnPropertySymbols){var s=Object.getOwnPropertySymbols(e);for(a=0;a<s.length;a++)n=s[a],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(i[n]=e[n])}return i}var d=a.createContext({}),p=function(e){var t=a.useContext(d),n=t;return e&&(n="function"==typeof e?e(t):r(r({},t),e)),n},o=function(e){var t=p(e.components);return a.createElement(d.Provider,{value:t},e.children)},u={inlineCode:"code",wrapper:function(e){var t=e.children;return a.createElement(a.Fragment,{},t)}},c=a.forwardRef((function(e,t){var n=e.components,i=e.mdxType,s=e.originalType,d=e.parentName,o=l(e,["components","mdxType","originalType","parentName"]),c=p(n),m=i,g=c["".concat(d,".").concat(m)]||c[m]||u[m]||s;return n?a.createElement(g,r(r({ref:t},o),{},{components:n})):a.createElement(g,r({ref:t},o))}));function m(e,t){var n=arguments,i=t&&t.mdxType;if("string"==typeof e||i){var s=n.length,r=new Array(s);r[0]=c;var l={};for(var d in t)hasOwnProperty.call(t,d)&&(l[d]=t[d]);l.originalType=e,l.mdxType="string"==typeof e?e:i,r[1]=l;for(var p=2;p<s;p++)r[p]=n[p];return a.createElement.apply(null,r)}return a.createElement.apply(null,n)}c.displayName="MDXCreateElement"},7531:function(e,t,n){n.r(t),n.d(t,{assets:function(){return o},contentTitle:function(){return d},default:function(){return m},frontMatter:function(){return l},metadata:function(){return p},toc:function(){return u}});var a=n(7462),i=n(3366),s=(n(7294),n(3905)),r=["components"],l={},d="Tagging",p={unversionedId:"guides/tagging",id:"guides/tagging",title:"Tagging",description:"This doc explains how to tag an asset in Compass with a specific tag.",source:"@site/docs/guides/tagging.md",sourceDirName:"guides",slug:"/guides/tagging",permalink:"/compass/guides/tagging",draft:!1,editUrl:"https://github.com/goto/compass/edit/master/docs/docs/guides/tagging.md",tags:[],version:"current",frontMatter:{},sidebar:"docsSidebar",previous:{title:"Starring",permalink:"/compass/guides/starring"},next:{title:"Discussion",permalink:"/compass/guides/discussion"}},o={},u=[{value:"Tag Template",id:"tag-template",level:2},{value:"Tagging an Asset",id:"tagging-an-asset",level:2},{value:"Getting Asset&#39;s Tag(s)",id:"getting-assets-tags",level:2}],c={toc:u};function m(e){var t=e.components,n=(0,i.Z)(e,r);return(0,s.kt)("wrapper",(0,a.Z)({},c,n,{components:t,mdxType:"MDXLayout"}),(0,s.kt)("h1",{id:"tagging"},"Tagging"),(0,s.kt)("p",null,"This doc explains how to tag an asset in Compass with a specific tag."),(0,s.kt)("h2",{id:"tag-template"},"Tag Template"),(0,s.kt)("p",null,"To support reusability of a tag, Compass has a tag template that we need to define first before we apply it to an asset. Tagging an asset means Compass will wire tag template to assets."),(0,s.kt)("p",null,"Creating a tag's template could be done with Tag Template API."),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-bash"},'$ curl --request POST \'localhost:8080/v1beta1/tags/templates\' \\\n--header \'Compass-User-UUID: user@gotocompany.com\' \\\n--data-raw \'{\n    "urn": "my-first-template",\n    "display_name": "My First Template",\n    "description": "This is my first template",\n    "fields": [\n        {\n            "urn": "fieldA",\n            "display_name": "Field A",\n            "description": "This is Field A",\n            "data_type": "string",\n            "required": false\n        },\n        {\n            "urn": "fieldB",\n            "display_name": "Field B",\n            "description": "This is Field B",\n            "data_type": "double",\n            "required": true\n        }\n    ]\n}\'\n')),(0,s.kt)("p",null,"We can verify the tag's template is created by calling GET tag's templates API"),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-bash"},"$ curl --request GET 'localhost:8080/v1beta1/tags/templates' \\\n--header 'Compass-User-UUID: user@gotocompany.com'\n")),(0,s.kt)("p",null,"The response will be like this"),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-javascript"},'{\n    "data": [\n        {\n            "urn": "my-first-template",\n            "display_name": "My First Template",\n            "description": "This is my first template",\n            "fields": [\n                {\n                    "id": 1,\n                    "urn": "fieldA",\n                    "display_name": "Field A",\n                    "description": "This is Field A",\n                    "data_type": "string",\n                    "created_at": "2022-05-10T09:34:18.766125Z",\n                    "updated_at": "2022-05-10T09:34:18.766125Z"\n                },\n                {\n                    "id": 2,\n                    "urn": "fieldB",\n                    "display_name": "Field B",\n                    "description": "This is Field B",\n                    "data_type": "double",\n                    "required": true,\n                    "created_at": "2022-05-10T09:34:18.766125Z",\n                    "updated_at": "2022-05-10T09:34:18.766125Z"\n                }\n            ],\n            "created_at": "2022-05-10T09:34:18.766125Z",\n            "updated_at": "2022-05-10T09:34:18.766125Z"\n        }\n    ]\n}\n')),(0,s.kt)("p",null,"Now, we already have a template with template urn ",(0,s.kt)("inlineCode",{parentName:"p"},"my-first-template")," that has 2 kind of fields with id ",(0,s.kt)("inlineCode",{parentName:"p"},"1")," and ",(0,s.kt)("inlineCode",{parentName:"p"},"2"),"."),(0,s.kt)("h2",{id:"tagging-an-asset"},"Tagging an Asset"),(0,s.kt)("p",null,"Once templates exist, we can tag an asset with a template by calling PUT ",(0,s.kt)("inlineCode",{parentName:"p"},"/v1beta1/tags/assets/{asset_id}")," API."),(0,s.kt)("p",null,"Assuming we have an asset"),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-javascript"},'{\n    "id": "a2c74793-b584-4d20-ba2a-28bdf6b92c08",\n    "urn": "sample-urn",\n    "type": "topic",\n    "service": "bigquery",\n    "name": "sample-name",\n    "description": "sample description",\n    "version": "0.1",\n    "updated_by": {\n        "uuid": "user@gotocompany.com"\n    },\n    "created_at": "2022-05-11T07:03:45.954387Z",\n    "updated_at": "2022-05-11T07:03:45.954387Z"\n}\n')),(0,s.kt)("p",null,"We can tag the asset with template ",(0,s.kt)("inlineCode",{parentName:"p"},"my-first-template"),"."),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-bash"},'$ curl --request POST \'localhost:8080/v1beta1/tags/assets\' \\\n--header \'Compass-User-UUID: user@gotocompany.com\'\n--data-raw \'{\n    "asset_id": "a2c74793-b584-4d20-ba2a-28bdf6b92c08",\n    "template_urn": "my-first-template",\n    "tag_values": [\n        {\n            "field_id": 1,\n            "field_value": "test"\n        },\n        {\n            "field_id": 2,\n            "field_value": 10.0\n        }\n    ]\n}\'\n')),(0,s.kt)("p",null,"We will get response showing that the asset is already tagged."),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-javascript"},'{\n    "data": {\n        "asset_id": "a2c74793-b584-4d20-ba2a-28bdf6b92c08",\n        "template_urn": "my-first-template",\n        "tag_values": [\n            {\n                "field_id": 1,\n                "field_value": "test",\n                "field_urn": "fieldA",\n                "field_display_name": "Field A",\n                "field_description": "This is Field A",\n                "field_data_type": "string",\n                "created_at": "2022-05-11T00:06:26.475943Z",\n                "updated_at": "2022-05-11T00:06:26.475943Z"\n            },\n            {\n                "field_id": 2,\n                "field_value": 10,\n                "field_urn": "fieldB",\n                "field_display_name": "Field B",\n                "field_description": "This is Field B",\n                "field_data_type": "double",\n                "field_required": true,\n                "created_at": "2022-05-11T00:06:26.475943Z",\n                "updated_at": "2022-05-11T00:06:26.475943Z"\n            }\n        ],\n        "template_display_name": "My First Template",\n        "template_description": "This is my first template"\n    }\n}\n')),(0,s.kt)("h2",{id:"getting-assets-tags"},"Getting Asset's Tag(s)"),(0,s.kt)("p",null,"We can get all tags belong to an asset by calling GET ",(0,s.kt)("inlineCode",{parentName:"p"},"/v1beta1/tags/assets/{asset_id}")," API."),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-bash"},'$ curl --request GET \'localhost:8080/v1beta1/tags/assets/a2c74793-b584-4d20-ba2a-28bdf6b92c08\' \\\n--header \'Compass-User-UUID: user@gotocompany.com\'\n\n{\n    "data": [\n        {\n            "asset_id": "a2c74793-b584-4d20-ba2a-28bdf6b92c08",\n            "template_urn": "my-first-template",\n            "tag_values": [\n                {\n                    "field_id": 1,\n                    "field_value": "test",\n                    "field_urn": "fieldA",\n                    "field_display_name": "Field A",\n                    "field_description": "This is Field A",\n                    "field_data_type": "string",\n                    "created_at": "2022-05-11T00:06:26.475943Z",\n                    "updated_at": "2022-05-11T00:06:26.475943Z"\n                },\n                {\n                    "field_id": 2,\n                    "field_value": 10,\n                    "field_urn": "fieldB",\n                    "field_display_name": "Field B",\n                    "field_description": "This is Field B",\n                    "field_data_type": "double",\n                    "field_required": true,\n                    "created_at": "2022-05-11T00:06:26.475943Z",\n                    "updated_at": "2022-05-11T00:06:26.475943Z"\n                }\n            ],\n            "template_display_name": "My First Template",\n            "template_description": "This is my first template"\n        }\n    ]\n}\n')))}m.isMDXComponent=!0}}]);