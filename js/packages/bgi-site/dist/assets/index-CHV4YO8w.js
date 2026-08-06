function Yv(r,e){for(var t=0;t<e.length;t++){const s=e[t];if(typeof s!="string"&&!Array.isArray(s)){for(const a in s)if(a!=="default"&&!(a in r)){const l=Object.getOwnPropertyDescriptor(s,a);l&&Object.defineProperty(r,a,l.get?l:{enumerable:!0,get:()=>s[a]})}}}return Object.freeze(Object.defineProperty(r,Symbol.toStringTag,{value:"Module"}))}(function(){const e=document.createElement("link").relList;if(e&&e.supports&&e.supports("modulepreload"))return;for(const a of document.querySelectorAll('link[rel="modulepreload"]'))s(a);new MutationObserver(a=>{for(const l of a)if(l.type==="childList")for(const u of l.addedNodes)u.tagName==="LINK"&&u.rel==="modulepreload"&&s(u)}).observe(document,{childList:!0,subtree:!0});function t(a){const l={};return a.integrity&&(l.integrity=a.integrity),a.referrerPolicy&&(l.referrerPolicy=a.referrerPolicy),a.crossOrigin==="use-credentials"?l.credentials="include":a.crossOrigin==="anonymous"?l.credentials="omit":l.credentials="same-origin",l}function s(a){if(a.ep)return;a.ep=!0;const l=t(a);fetch(a.href,l)}})();function qv(r){return r&&r.__esModule&&Object.prototype.hasOwnProperty.call(r,"default")?r.default:r}var Fu={exports:{}},Wo={},Ou={exports:{}},lt={};var Qp;function $v(){if(Qp)return lt;Qp=1;var r=Symbol.for("react.element"),e=Symbol.for("react.portal"),t=Symbol.for("react.fragment"),s=Symbol.for("react.strict_mode"),a=Symbol.for("react.profiler"),l=Symbol.for("react.provider"),u=Symbol.for("react.context"),f=Symbol.for("react.forward_ref"),h=Symbol.for("react.suspense"),p=Symbol.for("react.memo"),m=Symbol.for("react.lazy"),_=Symbol.iterator;function x(L){return L===null||typeof L!="object"?null:(L=_&&L[_]||L["@@iterator"],typeof L=="function"?L:null)}var S={isMounted:function(){return!1},enqueueForceUpdate:function(){},enqueueReplaceState:function(){},enqueueSetState:function(){}},T=Object.assign,w={};function v(L,X,ve){this.props=L,this.context=X,this.refs=w,this.updater=ve||S}v.prototype.isReactComponent={},v.prototype.setState=function(L,X){if(typeof L!="object"&&typeof L!="function"&&L!=null)throw Error("setState(...): takes an object of state variables to update or a function which returns an object of state variables.");this.updater.enqueueSetState(this,L,X,"setState")},v.prototype.forceUpdate=function(L){this.updater.enqueueForceUpdate(this,L,"forceUpdate")};function y(){}y.prototype=v.prototype;function P(L,X,ve){this.props=L,this.context=X,this.refs=w,this.updater=ve||S}var b=P.prototype=new y;b.constructor=P,T(b,v.prototype),b.isPureReactComponent=!0;var D=Array.isArray,V=Object.prototype.hasOwnProperty,O={current:null},U={key:!0,ref:!0,__self:!0,__source:!0};function Y(L,X,ve){var Ne,J={},fe=null,Se=null;if(X!=null)for(Ne in X.ref!==void 0&&(Se=X.ref),X.key!==void 0&&(fe=""+X.key),X)V.call(X,Ne)&&!U.hasOwnProperty(Ne)&&(J[Ne]=X[Ne]);var Me=arguments.length-2;if(Me===1)J.children=ve;else if(1<Me){for(var Pe=Array(Me),Ge=0;Ge<Me;Ge++)Pe[Ge]=arguments[Ge+2];J.children=Pe}if(L&&L.defaultProps)for(Ne in Me=L.defaultProps,Me)J[Ne]===void 0&&(J[Ne]=Me[Ne]);return{$$typeof:r,type:L,key:fe,ref:Se,props:J,_owner:O.current}}function ce(L,X){return{$$typeof:r,type:L.type,key:X,ref:L.ref,props:L.props,_owner:L._owner}}function E(L){return typeof L=="object"&&L!==null&&L.$$typeof===r}function C(L){var X={"=":"=0",":":"=2"};return"$"+L.replace(/[=:]/g,function(ve){return X[ve]})}var re=/\/+/g;function ee(L,X){return typeof L=="object"&&L!==null&&L.key!=null?C(""+L.key):X.toString(36)}function ae(L,X,ve,Ne,J){var fe=typeof L;(fe==="undefined"||fe==="boolean")&&(L=null);var Se=!1;if(L===null)Se=!0;else switch(fe){case"string":case"number":Se=!0;break;case"object":switch(L.$$typeof){case r:case e:Se=!0}}if(Se)return Se=L,J=J(Se),L=Ne===""?"."+ee(Se,0):Ne,D(J)?(ve="",L!=null&&(ve=L.replace(re,"$&/")+"/"),ae(J,X,ve,"",function(Ge){return Ge})):J!=null&&(E(J)&&(J=ce(J,ve+(!J.key||Se&&Se.key===J.key?"":(""+J.key).replace(re,"$&/")+"/")+L)),X.push(J)),1;if(Se=0,Ne=Ne===""?".":Ne+":",D(L))for(var Me=0;Me<L.length;Me++){fe=L[Me];var Pe=Ne+ee(fe,Me);Se+=ae(fe,X,ve,Pe,J)}else if(Pe=x(L),typeof Pe=="function")for(L=Pe.call(L),Me=0;!(fe=L.next()).done;)fe=fe.value,Pe=Ne+ee(fe,Me++),Se+=ae(fe,X,ve,Pe,J);else if(fe==="object")throw X=String(L),Error("Objects are not valid as a React child (found: "+(X==="[object Object]"?"object with keys {"+Object.keys(L).join(", ")+"}":X)+"). If you meant to render a collection of children, use an array instead.");return Se}function ue(L,X,ve){if(L==null)return L;var Ne=[],J=0;return ae(L,Ne,"","",function(fe){return X.call(ve,fe,J++)}),Ne}function Z(L){if(L._status===-1){var X=L._result;X=X(),X.then(function(ve){(L._status===0||L._status===-1)&&(L._status=1,L._result=ve)},function(ve){(L._status===0||L._status===-1)&&(L._status=2,L._result=ve)}),L._status===-1&&(L._status=0,L._result=X)}if(L._status===1)return L._result.default;throw L._result}var le={current:null},F={transition:null},se={ReactCurrentDispatcher:le,ReactCurrentBatchConfig:F,ReactCurrentOwner:O};return lt.Children={map:ue,forEach:function(L,X,ve){ue(L,function(){X.apply(this,arguments)},ve)},count:function(L){var X=0;return ue(L,function(){X++}),X},toArray:function(L){return ue(L,function(X){return X})||[]},only:function(L){if(!E(L))throw Error("React.Children.only expected to receive a single React element child.");return L}},lt.Component=v,lt.Fragment=t,lt.Profiler=a,lt.PureComponent=P,lt.StrictMode=s,lt.Suspense=h,lt.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED=se,lt.cloneElement=function(L,X,ve){if(L==null)throw Error("React.cloneElement(...): The argument must be a React element, but you passed "+L+".");var Ne=T({},L.props),J=L.key,fe=L.ref,Se=L._owner;if(X!=null){if(X.ref!==void 0&&(fe=X.ref,Se=O.current),X.key!==void 0&&(J=""+X.key),L.type&&L.type.defaultProps)var Me=L.type.defaultProps;for(Pe in X)V.call(X,Pe)&&!U.hasOwnProperty(Pe)&&(Ne[Pe]=X[Pe]===void 0&&Me!==void 0?Me[Pe]:X[Pe])}var Pe=arguments.length-2;if(Pe===1)Ne.children=ve;else if(1<Pe){Me=Array(Pe);for(var Ge=0;Ge<Pe;Ge++)Me[Ge]=arguments[Ge+2];Ne.children=Me}return{$$typeof:r,type:L.type,key:J,ref:fe,props:Ne,_owner:Se}},lt.createContext=function(L){return L={$$typeof:u,_currentValue:L,_currentValue2:L,_threadCount:0,Provider:null,Consumer:null,_defaultValue:null,_globalName:null},L.Provider={$$typeof:l,_context:L},L.Consumer=L},lt.createElement=Y,lt.createFactory=function(L){var X=Y.bind(null,L);return X.type=L,X},lt.createRef=function(){return{current:null}},lt.forwardRef=function(L){return{$$typeof:f,render:L}},lt.isValidElement=E,lt.lazy=function(L){return{$$typeof:m,_payload:{_status:-1,_result:L},_init:Z}},lt.memo=function(L,X){return{$$typeof:p,type:L,compare:X===void 0?null:X}},lt.startTransition=function(L){var X=F.transition;F.transition={};try{L()}finally{F.transition=X}},lt.unstable_act=function(){throw Error("act(...) is not supported in production builds of React.")},lt.useCallback=function(L,X){return le.current.useCallback(L,X)},lt.useContext=function(L){return le.current.useContext(L)},lt.useDebugValue=function(){},lt.useDeferredValue=function(L){return le.current.useDeferredValue(L)},lt.useEffect=function(L,X){return le.current.useEffect(L,X)},lt.useId=function(){return le.current.useId()},lt.useImperativeHandle=function(L,X,ve){return le.current.useImperativeHandle(L,X,ve)},lt.useInsertionEffect=function(L,X){return le.current.useInsertionEffect(L,X)},lt.useLayoutEffect=function(L,X){return le.current.useLayoutEffect(L,X)},lt.useMemo=function(L,X){return le.current.useMemo(L,X)},lt.useReducer=function(L,X,ve){return le.current.useReducer(L,X,ve)},lt.useRef=function(L){return le.current.useRef(L)},lt.useState=function(L){return le.current.useState(L)},lt.useSyncExternalStore=function(L,X,ve){return le.current.useSyncExternalStore(L,X,ve)},lt.useTransition=function(){return le.current.useTransition()},lt.version="18.2.0",lt}var Jp;function dd(){return Jp||(Jp=1,Ou.exports=$v()),Ou.exports}var em;function Kv(){if(em)return Wo;em=1;var r=dd(),e=Symbol.for("react.element"),t=Symbol.for("react.fragment"),s=Object.prototype.hasOwnProperty,a=r.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.ReactCurrentOwner,l={key:!0,ref:!0,__self:!0,__source:!0};function u(f,h,p){var m,_={},x=null,S=null;p!==void 0&&(x=""+p),h.key!==void 0&&(x=""+h.key),h.ref!==void 0&&(S=h.ref);for(m in h)s.call(h,m)&&!l.hasOwnProperty(m)&&(_[m]=h[m]);if(f&&f.defaultProps)for(m in h=f.defaultProps,h)_[m]===void 0&&(_[m]=h[m]);return{$$typeof:e,type:f,key:x,ref:S,props:_,_owner:a.current}}return Wo.Fragment=t,Wo.jsx=u,Wo.jsxs=u,Wo}var tm;function Zv(){return tm||(tm=1,Fu.exports=Kv()),Fu.exports}var k=Zv(),he=dd();const Qv=qv(he),Jv=Yv({__proto__:null,default:Qv},[he]);var hl={},ku={exports:{}},Rn={},Bu={exports:{}},zu={};var nm;function e0(){return nm||(nm=1,(function(r){function e(F,se){var L=F.length;F.push(se);e:for(;0<L;){var X=L-1>>>1,ve=F[X];if(0<a(ve,se))F[X]=se,F[L]=ve,L=X;else break e}}function t(F){return F.length===0?null:F[0]}function s(F){if(F.length===0)return null;var se=F[0],L=F.pop();if(L!==se){F[0]=L;e:for(var X=0,ve=F.length,Ne=ve>>>1;X<Ne;){var J=2*(X+1)-1,fe=F[J],Se=J+1,Me=F[Se];if(0>a(fe,L))Se<ve&&0>a(Me,fe)?(F[X]=Me,F[Se]=L,X=Se):(F[X]=fe,F[J]=L,X=J);else if(Se<ve&&0>a(Me,L))F[X]=Me,F[Se]=L,X=Se;else break e}}return se}function a(F,se){var L=F.sortIndex-se.sortIndex;return L!==0?L:F.id-se.id}if(typeof performance=="object"&&typeof performance.now=="function"){var l=performance;r.unstable_now=function(){return l.now()}}else{var u=Date,f=u.now();r.unstable_now=function(){return u.now()-f}}var h=[],p=[],m=1,_=null,x=3,S=!1,T=!1,w=!1,v=typeof setTimeout=="function"?setTimeout:null,y=typeof clearTimeout=="function"?clearTimeout:null,P=typeof setImmediate<"u"?setImmediate:null;typeof navigator<"u"&&navigator.scheduling!==void 0&&navigator.scheduling.isInputPending!==void 0&&navigator.scheduling.isInputPending.bind(navigator.scheduling);function b(F){for(var se=t(p);se!==null;){if(se.callback===null)s(p);else if(se.startTime<=F)s(p),se.sortIndex=se.expirationTime,e(h,se);else break;se=t(p)}}function D(F){if(w=!1,b(F),!T)if(t(h)!==null)T=!0,Z(V);else{var se=t(p);se!==null&&le(D,se.startTime-F)}}function V(F,se){T=!1,w&&(w=!1,y(Y),Y=-1),S=!0;var L=x;try{for(b(se),_=t(h);_!==null&&(!(_.expirationTime>se)||F&&!C());){var X=_.callback;if(typeof X=="function"){_.callback=null,x=_.priorityLevel;var ve=X(_.expirationTime<=se);se=r.unstable_now(),typeof ve=="function"?_.callback=ve:_===t(h)&&s(h),b(se)}else s(h);_=t(h)}if(_!==null)var Ne=!0;else{var J=t(p);J!==null&&le(D,J.startTime-se),Ne=!1}return Ne}finally{_=null,x=L,S=!1}}var O=!1,U=null,Y=-1,ce=5,E=-1;function C(){return!(r.unstable_now()-E<ce)}function re(){if(U!==null){var F=r.unstable_now();E=F;var se=!0;try{se=U(!0,F)}finally{se?ee():(O=!1,U=null)}}else O=!1}var ee;if(typeof P=="function")ee=function(){P(re)};else if(typeof MessageChannel<"u"){var ae=new MessageChannel,ue=ae.port2;ae.port1.onmessage=re,ee=function(){ue.postMessage(null)}}else ee=function(){v(re,0)};function Z(F){U=F,O||(O=!0,ee())}function le(F,se){Y=v(function(){F(r.unstable_now())},se)}r.unstable_IdlePriority=5,r.unstable_ImmediatePriority=1,r.unstable_LowPriority=4,r.unstable_NormalPriority=3,r.unstable_Profiling=null,r.unstable_UserBlockingPriority=2,r.unstable_cancelCallback=function(F){F.callback=null},r.unstable_continueExecution=function(){T||S||(T=!0,Z(V))},r.unstable_forceFrameRate=function(F){0>F||125<F?console.error("forceFrameRate takes a positive int between 0 and 125, forcing frame rates higher than 125 fps is not supported"):ce=0<F?Math.floor(1e3/F):5},r.unstable_getCurrentPriorityLevel=function(){return x},r.unstable_getFirstCallbackNode=function(){return t(h)},r.unstable_next=function(F){switch(x){case 1:case 2:case 3:var se=3;break;default:se=x}var L=x;x=se;try{return F()}finally{x=L}},r.unstable_pauseExecution=function(){},r.unstable_requestPaint=function(){},r.unstable_runWithPriority=function(F,se){switch(F){case 1:case 2:case 3:case 4:case 5:break;default:F=3}var L=x;x=F;try{return se()}finally{x=L}},r.unstable_scheduleCallback=function(F,se,L){var X=r.unstable_now();switch(typeof L=="object"&&L!==null?(L=L.delay,L=typeof L=="number"&&0<L?X+L:X):L=X,F){case 1:var ve=-1;break;case 2:ve=250;break;case 5:ve=1073741823;break;case 4:ve=1e4;break;default:ve=5e3}return ve=L+ve,F={id:m++,callback:se,priorityLevel:F,startTime:L,expirationTime:ve,sortIndex:-1},L>X?(F.sortIndex=L,e(p,F),t(h)===null&&F===t(p)&&(w?(y(Y),Y=-1):w=!0,le(D,L-X))):(F.sortIndex=ve,e(h,F),T||S||(T=!0,Z(V))),F},r.unstable_shouldYield=C,r.unstable_wrapCallback=function(F){var se=x;return function(){var L=x;x=se;try{return F.apply(this,arguments)}finally{x=L}}}})(zu)),zu}var im;function t0(){return im||(im=1,Bu.exports=e0()),Bu.exports}var rm;function n0(){if(rm)return Rn;rm=1;var r=dd(),e=t0();function t(n){for(var i="https://reactjs.org/docs/error-decoder.html?invariant="+n,o=1;o<arguments.length;o++)i+="&args[]="+encodeURIComponent(arguments[o]);return"Minified React error #"+n+"; visit "+i+" for the full message or use the non-minified dev environment for full errors and additional helpful warnings."}var s=new Set,a={};function l(n,i){u(n,i),u(n+"Capture",i)}function u(n,i){for(a[n]=i,n=0;n<i.length;n++)s.add(i[n])}var f=!(typeof window>"u"||typeof window.document>"u"||typeof window.document.createElement>"u"),h=Object.prototype.hasOwnProperty,p=/^[:A-Z_a-z\u00C0-\u00D6\u00D8-\u00F6\u00F8-\u02FF\u0370-\u037D\u037F-\u1FFF\u200C-\u200D\u2070-\u218F\u2C00-\u2FEF\u3001-\uD7FF\uF900-\uFDCF\uFDF0-\uFFFD][:A-Z_a-z\u00C0-\u00D6\u00D8-\u00F6\u00F8-\u02FF\u0370-\u037D\u037F-\u1FFF\u200C-\u200D\u2070-\u218F\u2C00-\u2FEF\u3001-\uD7FF\uF900-\uFDCF\uFDF0-\uFFFD\-.0-9\u00B7\u0300-\u036F\u203F-\u2040]*$/,m={},_={};function x(n){return h.call(_,n)?!0:h.call(m,n)?!1:p.test(n)?_[n]=!0:(m[n]=!0,!1)}function S(n,i,o,c){if(o!==null&&o.type===0)return!1;switch(typeof i){case"function":case"symbol":return!0;case"boolean":return c?!1:o!==null?!o.acceptsBooleans:(n=n.toLowerCase().slice(0,5),n!=="data-"&&n!=="aria-");default:return!1}}function T(n,i,o,c){if(i===null||typeof i>"u"||S(n,i,o,c))return!0;if(c)return!1;if(o!==null)switch(o.type){case 3:return!i;case 4:return i===!1;case 5:return isNaN(i);case 6:return isNaN(i)||1>i}return!1}function w(n,i,o,c,d,g,M){this.acceptsBooleans=i===2||i===3||i===4,this.attributeName=c,this.attributeNamespace=d,this.mustUseProperty=o,this.propertyName=n,this.type=i,this.sanitizeURL=g,this.removeEmptyString=M}var v={};"children dangerouslySetInnerHTML defaultValue defaultChecked innerHTML suppressContentEditableWarning suppressHydrationWarning style".split(" ").forEach(function(n){v[n]=new w(n,0,!1,n,null,!1,!1)}),[["acceptCharset","accept-charset"],["className","class"],["htmlFor","for"],["httpEquiv","http-equiv"]].forEach(function(n){var i=n[0];v[i]=new w(i,1,!1,n[1],null,!1,!1)}),["contentEditable","draggable","spellCheck","value"].forEach(function(n){v[n]=new w(n,2,!1,n.toLowerCase(),null,!1,!1)}),["autoReverse","externalResourcesRequired","focusable","preserveAlpha"].forEach(function(n){v[n]=new w(n,2,!1,n,null,!1,!1)}),"allowFullScreen async autoFocus autoPlay controls default defer disabled disablePictureInPicture disableRemotePlayback formNoValidate hidden loop noModule noValidate open playsInline readOnly required reversed scoped seamless itemScope".split(" ").forEach(function(n){v[n]=new w(n,3,!1,n.toLowerCase(),null,!1,!1)}),["checked","multiple","muted","selected"].forEach(function(n){v[n]=new w(n,3,!0,n,null,!1,!1)}),["capture","download"].forEach(function(n){v[n]=new w(n,4,!1,n,null,!1,!1)}),["cols","rows","size","span"].forEach(function(n){v[n]=new w(n,6,!1,n,null,!1,!1)}),["rowSpan","start"].forEach(function(n){v[n]=new w(n,5,!1,n.toLowerCase(),null,!1,!1)});var y=/[\-:]([a-z])/g;function P(n){return n[1].toUpperCase()}"accent-height alignment-baseline arabic-form baseline-shift cap-height clip-path clip-rule color-interpolation color-interpolation-filters color-profile color-rendering dominant-baseline enable-background fill-opacity fill-rule flood-color flood-opacity font-family font-size font-size-adjust font-stretch font-style font-variant font-weight glyph-name glyph-orientation-horizontal glyph-orientation-vertical horiz-adv-x horiz-origin-x image-rendering letter-spacing lighting-color marker-end marker-mid marker-start overline-position overline-thickness paint-order panose-1 pointer-events rendering-intent shape-rendering stop-color stop-opacity strikethrough-position strikethrough-thickness stroke-dasharray stroke-dashoffset stroke-linecap stroke-linejoin stroke-miterlimit stroke-opacity stroke-width text-anchor text-decoration text-rendering underline-position underline-thickness unicode-bidi unicode-range units-per-em v-alphabetic v-hanging v-ideographic v-mathematical vector-effect vert-adv-y vert-origin-x vert-origin-y word-spacing writing-mode xmlns:xlink x-height".split(" ").forEach(function(n){var i=n.replace(y,P);v[i]=new w(i,1,!1,n,null,!1,!1)}),"xlink:actuate xlink:arcrole xlink:role xlink:show xlink:title xlink:type".split(" ").forEach(function(n){var i=n.replace(y,P);v[i]=new w(i,1,!1,n,"http://www.w3.org/1999/xlink",!1,!1)}),["xml:base","xml:lang","xml:space"].forEach(function(n){var i=n.replace(y,P);v[i]=new w(i,1,!1,n,"http://www.w3.org/XML/1998/namespace",!1,!1)}),["tabIndex","crossOrigin"].forEach(function(n){v[n]=new w(n,1,!1,n.toLowerCase(),null,!1,!1)}),v.xlinkHref=new w("xlinkHref",1,!1,"xlink:href","http://www.w3.org/1999/xlink",!0,!1),["src","href","action","formAction"].forEach(function(n){v[n]=new w(n,1,!1,n.toLowerCase(),null,!0,!0)});function b(n,i,o,c){var d=v.hasOwnProperty(i)?v[i]:null;(d!==null?d.type!==0:c||!(2<i.length)||i[0]!=="o"&&i[0]!=="O"||i[1]!=="n"&&i[1]!=="N")&&(T(i,o,d,c)&&(o=null),c||d===null?x(i)&&(o===null?n.removeAttribute(i):n.setAttribute(i,""+o)):d.mustUseProperty?n[d.propertyName]=o===null?d.type===3?!1:"":o:(i=d.attributeName,c=d.attributeNamespace,o===null?n.removeAttribute(i):(d=d.type,o=d===3||d===4&&o===!0?"":""+o,c?n.setAttributeNS(c,i,o):n.setAttribute(i,o))))}var D=r.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED,V=Symbol.for("react.element"),O=Symbol.for("react.portal"),U=Symbol.for("react.fragment"),Y=Symbol.for("react.strict_mode"),ce=Symbol.for("react.profiler"),E=Symbol.for("react.provider"),C=Symbol.for("react.context"),re=Symbol.for("react.forward_ref"),ee=Symbol.for("react.suspense"),ae=Symbol.for("react.suspense_list"),ue=Symbol.for("react.memo"),Z=Symbol.for("react.lazy"),le=Symbol.for("react.offscreen"),F=Symbol.iterator;function se(n){return n===null||typeof n!="object"?null:(n=F&&n[F]||n["@@iterator"],typeof n=="function"?n:null)}var L=Object.assign,X;function ve(n){if(X===void 0)try{throw Error()}catch(o){var i=o.stack.trim().match(/\n( *(at )?)/);X=i&&i[1]||""}return`
`+X+n}var Ne=!1;function J(n,i){if(!n||Ne)return"";Ne=!0;var o=Error.prepareStackTrace;Error.prepareStackTrace=void 0;try{if(i)if(i=function(){throw Error()},Object.defineProperty(i.prototype,"props",{set:function(){throw Error()}}),typeof Reflect=="object"&&Reflect.construct){try{Reflect.construct(i,[])}catch(ie){var c=ie}Reflect.construct(n,[],i)}else{try{i.call()}catch(ie){c=ie}n.call(i.prototype)}else{try{throw Error()}catch(ie){c=ie}n()}}catch(ie){if(ie&&c&&typeof ie.stack=="string"){for(var d=ie.stack.split(`
`),g=c.stack.split(`
`),M=d.length-1,I=g.length-1;1<=M&&0<=I&&d[M]!==g[I];)I--;for(;1<=M&&0<=I;M--,I--)if(d[M]!==g[I]){if(M!==1||I!==1)do if(M--,I--,0>I||d[M]!==g[I]){var B=`
`+d[M].replace(" at new "," at ");return n.displayName&&B.includes("<anonymous>")&&(B=B.replace("<anonymous>",n.displayName)),B}while(1<=M&&0<=I);break}}}finally{Ne=!1,Error.prepareStackTrace=o}return(n=n?n.displayName||n.name:"")?ve(n):""}function fe(n){switch(n.tag){case 5:return ve(n.type);case 16:return ve("Lazy");case 13:return ve("Suspense");case 19:return ve("SuspenseList");case 0:case 2:case 15:return n=J(n.type,!1),n;case 11:return n=J(n.type.render,!1),n;case 1:return n=J(n.type,!0),n;default:return""}}function Se(n){if(n==null)return null;if(typeof n=="function")return n.displayName||n.name||null;if(typeof n=="string")return n;switch(n){case U:return"Fragment";case O:return"Portal";case ce:return"Profiler";case Y:return"StrictMode";case ee:return"Suspense";case ae:return"SuspenseList"}if(typeof n=="object")switch(n.$$typeof){case C:return(n.displayName||"Context")+".Consumer";case E:return(n._context.displayName||"Context")+".Provider";case re:var i=n.render;return n=n.displayName,n||(n=i.displayName||i.name||"",n=n!==""?"ForwardRef("+n+")":"ForwardRef"),n;case ue:return i=n.displayName||null,i!==null?i:Se(n.type)||"Memo";case Z:i=n._payload,n=n._init;try{return Se(n(i))}catch{}}return null}function Me(n){var i=n.type;switch(n.tag){case 24:return"Cache";case 9:return(i.displayName||"Context")+".Consumer";case 10:return(i._context.displayName||"Context")+".Provider";case 18:return"DehydratedFragment";case 11:return n=i.render,n=n.displayName||n.name||"",i.displayName||(n!==""?"ForwardRef("+n+")":"ForwardRef");case 7:return"Fragment";case 5:return i;case 4:return"Portal";case 3:return"Root";case 6:return"Text";case 16:return Se(i);case 8:return i===Y?"StrictMode":"Mode";case 22:return"Offscreen";case 12:return"Profiler";case 21:return"Scope";case 13:return"Suspense";case 19:return"SuspenseList";case 25:return"TracingMarker";case 1:case 0:case 17:case 2:case 14:case 15:if(typeof i=="function")return i.displayName||i.name||null;if(typeof i=="string")return i}return null}function Pe(n){switch(typeof n){case"boolean":case"number":case"string":case"undefined":return n;case"object":return n;default:return""}}function Ge(n){var i=n.type;return(n=n.nodeName)&&n.toLowerCase()==="input"&&(i==="checkbox"||i==="radio")}function dt(n){var i=Ge(n)?"checked":"value",o=Object.getOwnPropertyDescriptor(n.constructor.prototype,i),c=""+n[i];if(!n.hasOwnProperty(i)&&typeof o<"u"&&typeof o.get=="function"&&typeof o.set=="function"){var d=o.get,g=o.set;return Object.defineProperty(n,i,{configurable:!0,get:function(){return d.call(this)},set:function(M){c=""+M,g.call(this,M)}}),Object.defineProperty(n,i,{enumerable:o.enumerable}),{getValue:function(){return c},setValue:function(M){c=""+M},stopTracking:function(){n._valueTracker=null,delete n[i]}}}}function gt(n){n._valueTracker||(n._valueTracker=dt(n))}function ct(n){if(!n)return!1;var i=n._valueTracker;if(!i)return!0;var o=i.getValue(),c="";return n&&(c=Ge(n)?n.checked?"true":"false":n.value),n=c,n!==o?(i.setValue(n),!0):!1}function z(n){if(n=n||(typeof document<"u"?document:void 0),typeof n>"u")return null;try{return n.activeElement||n.body}catch{return n.body}}function an(n,i){var o=i.checked;return L({},i,{defaultChecked:void 0,defaultValue:void 0,value:void 0,checked:o??n._wrapperState.initialChecked})}function at(n,i){var o=i.defaultValue==null?"":i.defaultValue,c=i.checked!=null?i.checked:i.defaultChecked;o=Pe(i.value!=null?i.value:o),n._wrapperState={initialChecked:c,initialValue:o,controlled:i.type==="checkbox"||i.type==="radio"?i.checked!=null:i.value!=null}}function ht(n,i){i=i.checked,i!=null&&b(n,"checked",i,!1)}function Ze(n,i){ht(n,i);var o=Pe(i.value),c=i.type;if(o!=null)c==="number"?(o===0&&n.value===""||n.value!=o)&&(n.value=""+o):n.value!==""+o&&(n.value=""+o);else if(c==="submit"||c==="reset"){n.removeAttribute("value");return}i.hasOwnProperty("value")?Je(n,i.type,o):i.hasOwnProperty("defaultValue")&&Je(n,i.type,Pe(i.defaultValue)),i.checked==null&&i.defaultChecked!=null&&(n.defaultChecked=!!i.defaultChecked)}function At(n,i,o){if(i.hasOwnProperty("value")||i.hasOwnProperty("defaultValue")){var c=i.type;if(!(c!=="submit"&&c!=="reset"||i.value!==void 0&&i.value!==null))return;i=""+n._wrapperState.initialValue,o||i===n.value||(n.value=i),n.defaultValue=i}o=n.name,o!==""&&(n.name=""),n.defaultChecked=!!n._wrapperState.initialChecked,o!==""&&(n.name=o)}function Je(n,i,o){(i!=="number"||z(n.ownerDocument)!==n)&&(o==null?n.defaultValue=""+n._wrapperState.initialValue:n.defaultValue!==""+o&&(n.defaultValue=""+o))}var N=Array.isArray;function A(n,i,o,c){if(n=n.options,i){i={};for(var d=0;d<o.length;d++)i["$"+o[d]]=!0;for(o=0;o<n.length;o++)d=i.hasOwnProperty("$"+n[o].value),n[o].selected!==d&&(n[o].selected=d),d&&c&&(n[o].defaultSelected=!0)}else{for(o=""+Pe(o),i=null,d=0;d<n.length;d++){if(n[d].value===o){n[d].selected=!0,c&&(n[d].defaultSelected=!0);return}i!==null||n[d].disabled||(i=n[d])}i!==null&&(i.selected=!0)}}function K(n,i){if(i.dangerouslySetInnerHTML!=null)throw Error(t(91));return L({},i,{value:void 0,defaultValue:void 0,children:""+n._wrapperState.initialValue})}function pe(n,i){var o=i.value;if(o==null){if(o=i.children,i=i.defaultValue,o!=null){if(i!=null)throw Error(t(92));if(N(o)){if(1<o.length)throw Error(t(93));o=o[0]}i=o}i==null&&(i=""),o=i}n._wrapperState={initialValue:Pe(o)}}function xe(n,i){var o=Pe(i.value),c=Pe(i.defaultValue);o!=null&&(o=""+o,o!==n.value&&(n.value=o),i.defaultValue==null&&n.defaultValue!==o&&(n.defaultValue=o)),c!=null&&(n.defaultValue=""+c)}function de(n){var i=n.textContent;i===n._wrapperState.initialValue&&i!==""&&i!==null&&(n.value=i)}function Ye(n){switch(n){case"svg":return"http://www.w3.org/2000/svg";case"math":return"http://www.w3.org/1998/Math/MathML";default:return"http://www.w3.org/1999/xhtml"}}function Ce(n,i){return n==null||n==="http://www.w3.org/1999/xhtml"?Ye(i):n==="http://www.w3.org/2000/svg"&&i==="foreignObject"?"http://www.w3.org/1999/xhtml":n}var Fe,pt=(function(n){return typeof MSApp<"u"&&MSApp.execUnsafeLocalFunction?function(i,o,c,d){MSApp.execUnsafeLocalFunction(function(){return n(i,o,c,d)})}:n})(function(n,i){if(n.namespaceURI!=="http://www.w3.org/2000/svg"||"innerHTML"in n)n.innerHTML=i;else{for(Fe=Fe||document.createElement("div"),Fe.innerHTML="<svg>"+i.valueOf().toString()+"</svg>",i=Fe.firstChild;n.firstChild;)n.removeChild(n.firstChild);for(;i.firstChild;)n.appendChild(i.firstChild)}});function Ee(n,i){if(i){var o=n.firstChild;if(o&&o===n.lastChild&&o.nodeType===3){o.nodeValue=i;return}}n.textContent=i}var Oe={animationIterationCount:!0,aspectRatio:!0,borderImageOutset:!0,borderImageSlice:!0,borderImageWidth:!0,boxFlex:!0,boxFlexGroup:!0,boxOrdinalGroup:!0,columnCount:!0,columns:!0,flex:!0,flexGrow:!0,flexPositive:!0,flexShrink:!0,flexNegative:!0,flexOrder:!0,gridArea:!0,gridRow:!0,gridRowEnd:!0,gridRowSpan:!0,gridRowStart:!0,gridColumn:!0,gridColumnEnd:!0,gridColumnSpan:!0,gridColumnStart:!0,fontWeight:!0,lineClamp:!0,lineHeight:!0,opacity:!0,order:!0,orphans:!0,tabSize:!0,widows:!0,zIndex:!0,zoom:!0,fillOpacity:!0,floodOpacity:!0,stopOpacity:!0,strokeDasharray:!0,strokeDashoffset:!0,strokeMiterlimit:!0,strokeOpacity:!0,strokeWidth:!0},tt=["Webkit","ms","Moz","O"];Object.keys(Oe).forEach(function(n){tt.forEach(function(i){i=i+n.charAt(0).toUpperCase()+n.substring(1),Oe[i]=Oe[n]})});function et(n,i,o){return i==null||typeof i=="boolean"||i===""?"":o||typeof i!="number"||i===0||Oe.hasOwnProperty(n)&&Oe[n]?(""+i).trim():i+"px"}function Be(n,i){n=n.style;for(var o in i)if(i.hasOwnProperty(o)){var c=o.indexOf("--")===0,d=et(o,i[o],c);o==="float"&&(o="cssFloat"),c?n.setProperty(o,d):n[o]=d}}var ut=L({menuitem:!0},{area:!0,base:!0,br:!0,col:!0,embed:!0,hr:!0,img:!0,input:!0,keygen:!0,link:!0,meta:!0,param:!0,source:!0,track:!0,wbr:!0});function it(n,i){if(i){if(ut[n]&&(i.children!=null||i.dangerouslySetInnerHTML!=null))throw Error(t(137,n));if(i.dangerouslySetInnerHTML!=null){if(i.children!=null)throw Error(t(60));if(typeof i.dangerouslySetInnerHTML!="object"||!("__html"in i.dangerouslySetInnerHTML))throw Error(t(61))}if(i.style!=null&&typeof i.style!="object")throw Error(t(62))}}function Et(n,i){if(n.indexOf("-")===-1)return typeof i.is=="string";switch(n){case"annotation-xml":case"color-profile":case"font-face":case"font-face-src":case"font-face-uri":case"font-face-format":case"font-face-name":case"missing-glyph":return!1;default:return!0}}var G=null;function Le(n){return n=n.target||n.srcElement||window,n.correspondingUseElement&&(n=n.correspondingUseElement),n.nodeType===3?n.parentNode:n}var oe=null,me=null,Re=null;function Ie(n){if(n=bo(n)){if(typeof oe!="function")throw Error(t(280));var i=n.stateNode;i&&(i=Ra(i),oe(n.stateNode,n.type,i))}}function ft(n){me?Re?Re.push(n):Re=[n]:me=n}function kt(){if(me){var n=me,i=Re;if(Re=me=null,Ie(n),i)for(n=0;n<i.length;n++)Ie(i[n])}}function ln(n,i){return n(i)}function mt(){}var en=!1;function Gn(n,i,o){if(en)return n(i,o);en=!0;try{return ln(n,i,o)}finally{en=!1,(me!==null||Re!==null)&&(mt(),kt())}}function Xi(n,i){var o=n.stateNode;if(o===null)return null;var c=Ra(o);if(c===null)return null;o=c[i];e:switch(i){case"onClick":case"onClickCapture":case"onDoubleClick":case"onDoubleClickCapture":case"onMouseDown":case"onMouseDownCapture":case"onMouseMove":case"onMouseMoveCapture":case"onMouseUp":case"onMouseUpCapture":case"onMouseEnter":(c=!c.disabled)||(n=n.type,c=!(n==="button"||n==="input"||n==="select"||n==="textarea")),n=!c;break e;default:n=!1}if(n)return null;if(o&&typeof o!="function")throw Error(t(231,i,typeof o));return o}var as=!1;if(f)try{var Nn={};Object.defineProperty(Nn,"passive",{get:function(){as=!0}}),window.addEventListener("test",Nn,Nn),window.removeEventListener("test",Nn,Nn)}catch{as=!1}function co(n,i,o,c,d,g,M,I,B){var ie=Array.prototype.slice.call(arguments,3);try{i.apply(o,ie)}catch(_e){this.onError(_e)}}var Yi=!1,Pr=null,Mi=!1,ls=null,cs={onError:function(n){Yi=!0,Pr=n}};function ca(n,i,o,c,d,g,M,I,B){Yi=!1,Pr=null,co.apply(cs,arguments)}function ua(n,i,o,c,d,g,M,I,B){if(ca.apply(this,arguments),Yi){if(Yi){var ie=Pr;Yi=!1,Pr=null}else throw Error(t(198));Mi||(Mi=!0,ls=ie)}}function Ei(n){var i=n,o=n;if(n.alternate)for(;i.return;)i=i.return;else{n=i;do i=n,(i.flags&4098)!==0&&(o=i.return),n=i.return;while(n)}return i.tag===3?o:null}function fa(n){if(n.tag===13){var i=n.memoizedState;if(i===null&&(n=n.alternate,n!==null&&(i=n.memoizedState)),i!==null)return i.dehydrated}return null}function da(n){if(Ei(n)!==n)throw Error(t(188))}function R(n){var i=n.alternate;if(!i){if(i=Ei(n),i===null)throw Error(t(188));return i!==n?null:n}for(var o=n,c=i;;){var d=o.return;if(d===null)break;var g=d.alternate;if(g===null){if(c=d.return,c!==null){o=c;continue}break}if(d.child===g.child){for(g=d.child;g;){if(g===o)return da(d),n;if(g===c)return da(d),i;g=g.sibling}throw Error(t(188))}if(o.return!==c.return)o=d,c=g;else{for(var M=!1,I=d.child;I;){if(I===o){M=!0,o=d,c=g;break}if(I===c){M=!0,c=d,o=g;break}I=I.sibling}if(!M){for(I=g.child;I;){if(I===o){M=!0,o=g,c=d;break}if(I===c){M=!0,c=g,o=d;break}I=I.sibling}if(!M)throw Error(t(189))}}if(o.alternate!==c)throw Error(t(190))}if(o.tag!==3)throw Error(t(188));return o.stateNode.current===o?n:i}function W(n){return n=R(n),n!==null?te(n):null}function te(n){if(n.tag===5||n.tag===6)return n;for(n=n.child;n!==null;){var i=te(n);if(i!==null)return i;n=n.sibling}return null}var ne=e.unstable_scheduleCallback,j=e.unstable_cancelCallback,we=e.unstable_shouldYield,De=e.unstable_requestPaint,Ae=e.unstable_now,We=e.unstable_getCurrentPriorityLevel,Ke=e.unstable_ImmediatePriority,Qe=e.unstable_UserBlockingPriority,je=e.unstable_NormalPriority,Mt=e.unstable_LowPriority,Ct=e.unstable_IdlePriority,bt=null,Ft=null;function xt(n){if(Ft&&typeof Ft.onCommitFiberRoot=="function")try{Ft.onCommitFiberRoot(bt,n,void 0,(n.current.flags&128)===128)}catch{}}var ke=Math.clz32?Math.clz32:In,qt=Math.log,yt=Math.LN2;function In(n){return n>>>=0,n===0?32:31-(qt(n)/yt|0)|0}var Jn=64,tn=4194304;function Ti(n){switch(n&-n){case 1:return 1;case 2:return 2;case 4:return 4;case 8:return 8;case 16:return 16;case 32:return 32;case 64:case 128:case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:case 262144:case 524288:case 1048576:case 2097152:return n&4194240;case 4194304:case 8388608:case 16777216:case 33554432:case 67108864:return n&130023424;case 134217728:return 134217728;case 268435456:return 268435456;case 536870912:return 536870912;case 1073741824:return 1073741824;default:return n}}function Lt(n,i){var o=n.pendingLanes;if(o===0)return 0;var c=0,d=n.suspendedLanes,g=n.pingedLanes,M=o&268435455;if(M!==0){var I=M&~d;I!==0?c=Ti(I):(g&=M,g!==0&&(c=Ti(g)))}else M=o&~d,M!==0?c=Ti(M):g!==0&&(c=Ti(g));if(c===0)return 0;if(i!==0&&i!==c&&(i&d)===0&&(d=c&-c,g=i&-i,d>=g||d===16&&(g&4194240)!==0))return i;if((c&4)!==0&&(c|=o&16),i=n.entangledLanes,i!==0)for(n=n.entanglements,i&=c;0<i;)o=31-ke(i),d=1<<o,c|=n[o],i&=~d;return c}function hi(n,i){switch(n){case 1:case 2:case 4:return i+250;case 8:case 16:case 32:case 64:case 128:case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:case 262144:case 524288:case 1048576:case 2097152:return i+5e3;case 4194304:case 8388608:case 16777216:case 33554432:case 67108864:return-1;case 134217728:case 268435456:case 536870912:case 1073741824:return-1;default:return-1}}function uo(n,i){for(var o=n.suspendedLanes,c=n.pingedLanes,d=n.expirationTimes,g=n.pendingLanes;0<g;){var M=31-ke(g),I=1<<M,B=d[M];B===-1?((I&o)===0||(I&c)!==0)&&(d[M]=hi(I,i)):B<=i&&(n.expiredLanes|=I),g&=~I}}function fn(n){return n=n.pendingLanes&-1073741825,n!==0?n:n&1073741824?1073741824:0}function us(){var n=Jn;return Jn<<=1,(Jn&4194240)===0&&(Jn=64),n}function fo(n){for(var i=[],o=0;31>o;o++)i.push(n);return i}function qi(n,i,o){n.pendingLanes|=i,i!==536870912&&(n.suspendedLanes=0,n.pingedLanes=0),n=n.eventTimes,i=31-ke(i),n[i]=o}function p_(n,i){var o=n.pendingLanes&~i;n.pendingLanes=i,n.suspendedLanes=0,n.pingedLanes=0,n.expiredLanes&=i,n.mutableReadLanes&=i,n.entangledLanes&=i,i=n.entanglements;var c=n.eventTimes;for(n=n.expirationTimes;0<o;){var d=31-ke(o),g=1<<d;i[d]=0,c[d]=-1,n[d]=-1,o&=~g}}function sc(n,i){var o=n.entangledLanes|=i;for(n=n.entanglements;o;){var c=31-ke(o),d=1<<c;d&i|n[c]&i&&(n[c]|=i),o&=~d}}var wt=0;function bd(n){return n&=-n,1<n?4<n?(n&268435455)!==0?16:536870912:4:1}var Pd,oc,Ld,Dd,Nd,ac=!1,ha=[],$i=null,Ki=null,Zi=null,ho=new Map,po=new Map,Qi=[],m_="mousedown mouseup touchcancel touchend touchstart auxclick dblclick pointercancel pointerdown pointerup dragend dragstart drop compositionend compositionstart keydown keypress keyup input textInput copy cut paste click change contextmenu reset submit".split(" ");function Id(n,i){switch(n){case"focusin":case"focusout":$i=null;break;case"dragenter":case"dragleave":Ki=null;break;case"mouseover":case"mouseout":Zi=null;break;case"pointerover":case"pointerout":ho.delete(i.pointerId);break;case"gotpointercapture":case"lostpointercapture":po.delete(i.pointerId)}}function mo(n,i,o,c,d,g){return n===null||n.nativeEvent!==g?(n={blockedOn:i,domEventName:o,eventSystemFlags:c,nativeEvent:g,targetContainers:[d]},i!==null&&(i=bo(i),i!==null&&oc(i)),n):(n.eventSystemFlags|=c,i=n.targetContainers,d!==null&&i.indexOf(d)===-1&&i.push(d),n)}function g_(n,i,o,c,d){switch(i){case"focusin":return $i=mo($i,n,i,o,c,d),!0;case"dragenter":return Ki=mo(Ki,n,i,o,c,d),!0;case"mouseover":return Zi=mo(Zi,n,i,o,c,d),!0;case"pointerover":var g=d.pointerId;return ho.set(g,mo(ho.get(g)||null,n,i,o,c,d)),!0;case"gotpointercapture":return g=d.pointerId,po.set(g,mo(po.get(g)||null,n,i,o,c,d)),!0}return!1}function Ud(n){var i=Lr(n.target);if(i!==null){var o=Ei(i);if(o!==null){if(i=o.tag,i===13){if(i=fa(o),i!==null){n.blockedOn=i,Nd(n.priority,function(){Ld(o)});return}}else if(i===3&&o.stateNode.current.memoizedState.isDehydrated){n.blockedOn=o.tag===3?o.stateNode.containerInfo:null;return}}}n.blockedOn=null}function pa(n){if(n.blockedOn!==null)return!1;for(var i=n.targetContainers;0<i.length;){var o=cc(n.domEventName,n.eventSystemFlags,i[0],n.nativeEvent);if(o===null){o=n.nativeEvent;var c=new o.constructor(o.type,o);G=c,o.target.dispatchEvent(c),G=null}else return i=bo(o),i!==null&&oc(i),n.blockedOn=o,!1;i.shift()}return!0}function Fd(n,i,o){pa(n)&&o.delete(i)}function __(){ac=!1,$i!==null&&pa($i)&&($i=null),Ki!==null&&pa(Ki)&&(Ki=null),Zi!==null&&pa(Zi)&&(Zi=null),ho.forEach(Fd),po.forEach(Fd)}function go(n,i){n.blockedOn===i&&(n.blockedOn=null,ac||(ac=!0,e.unstable_scheduleCallback(e.unstable_NormalPriority,__)))}function _o(n){function i(d){return go(d,n)}if(0<ha.length){go(ha[0],n);for(var o=1;o<ha.length;o++){var c=ha[o];c.blockedOn===n&&(c.blockedOn=null)}}for($i!==null&&go($i,n),Ki!==null&&go(Ki,n),Zi!==null&&go(Zi,n),ho.forEach(i),po.forEach(i),o=0;o<Qi.length;o++)c=Qi[o],c.blockedOn===n&&(c.blockedOn=null);for(;0<Qi.length&&(o=Qi[0],o.blockedOn===null);)Ud(o),o.blockedOn===null&&Qi.shift()}var fs=D.ReactCurrentBatchConfig,ma=!0;function v_(n,i,o,c){var d=wt,g=fs.transition;fs.transition=null;try{wt=1,lc(n,i,o,c)}finally{wt=d,fs.transition=g}}function x_(n,i,o,c){var d=wt,g=fs.transition;fs.transition=null;try{wt=4,lc(n,i,o,c)}finally{wt=d,fs.transition=g}}function lc(n,i,o,c){if(ma){var d=cc(n,i,o,c);if(d===null)Ac(n,i,c,ga,o),Id(n,c);else if(g_(d,n,i,o,c))c.stopPropagation();else if(Id(n,c),i&4&&-1<m_.indexOf(n)){for(;d!==null;){var g=bo(d);if(g!==null&&Pd(g),g=cc(n,i,o,c),g===null&&Ac(n,i,c,ga,o),g===d)break;d=g}d!==null&&c.stopPropagation()}else Ac(n,i,c,null,o)}}var ga=null;function cc(n,i,o,c){if(ga=null,n=Le(c),n=Lr(n),n!==null)if(i=Ei(n),i===null)n=null;else if(o=i.tag,o===13){if(n=fa(i),n!==null)return n;n=null}else if(o===3){if(i.stateNode.current.memoizedState.isDehydrated)return i.tag===3?i.stateNode.containerInfo:null;n=null}else i!==n&&(n=null);return ga=n,null}function Od(n){switch(n){case"cancel":case"click":case"close":case"contextmenu":case"copy":case"cut":case"auxclick":case"dblclick":case"dragend":case"dragstart":case"drop":case"focusin":case"focusout":case"input":case"invalid":case"keydown":case"keypress":case"keyup":case"mousedown":case"mouseup":case"paste":case"pause":case"play":case"pointercancel":case"pointerdown":case"pointerup":case"ratechange":case"reset":case"resize":case"seeked":case"submit":case"touchcancel":case"touchend":case"touchstart":case"volumechange":case"change":case"selectionchange":case"textInput":case"compositionstart":case"compositionend":case"compositionupdate":case"beforeblur":case"afterblur":case"beforeinput":case"blur":case"fullscreenchange":case"focus":case"hashchange":case"popstate":case"select":case"selectstart":return 1;case"drag":case"dragenter":case"dragexit":case"dragleave":case"dragover":case"mousemove":case"mouseout":case"mouseover":case"pointermove":case"pointerout":case"pointerover":case"scroll":case"toggle":case"touchmove":case"wheel":case"mouseenter":case"mouseleave":case"pointerenter":case"pointerleave":return 4;case"message":switch(We()){case Ke:return 1;case Qe:return 4;case je:case Mt:return 16;case Ct:return 536870912;default:return 16}default:return 16}}var Ji=null,uc=null,_a=null;function kd(){if(_a)return _a;var n,i=uc,o=i.length,c,d="value"in Ji?Ji.value:Ji.textContent,g=d.length;for(n=0;n<o&&i[n]===d[n];n++);var M=o-n;for(c=1;c<=M&&i[o-c]===d[g-c];c++);return _a=d.slice(n,1<c?1-c:void 0)}function va(n){var i=n.keyCode;return"charCode"in n?(n=n.charCode,n===0&&i===13&&(n=13)):n=i,n===10&&(n=13),32<=n||n===13?n:0}function xa(){return!0}function Bd(){return!1}function Un(n){function i(o,c,d,g,M){this._reactName=o,this._targetInst=d,this.type=c,this.nativeEvent=g,this.target=M,this.currentTarget=null;for(var I in n)n.hasOwnProperty(I)&&(o=n[I],this[I]=o?o(g):g[I]);return this.isDefaultPrevented=(g.defaultPrevented!=null?g.defaultPrevented:g.returnValue===!1)?xa:Bd,this.isPropagationStopped=Bd,this}return L(i.prototype,{preventDefault:function(){this.defaultPrevented=!0;var o=this.nativeEvent;o&&(o.preventDefault?o.preventDefault():typeof o.returnValue!="unknown"&&(o.returnValue=!1),this.isDefaultPrevented=xa)},stopPropagation:function(){var o=this.nativeEvent;o&&(o.stopPropagation?o.stopPropagation():typeof o.cancelBubble!="unknown"&&(o.cancelBubble=!0),this.isPropagationStopped=xa)},persist:function(){},isPersistent:xa}),i}var ds={eventPhase:0,bubbles:0,cancelable:0,timeStamp:function(n){return n.timeStamp||Date.now()},defaultPrevented:0,isTrusted:0},fc=Un(ds),vo=L({},ds,{view:0,detail:0}),y_=Un(vo),dc,hc,xo,ya=L({},vo,{screenX:0,screenY:0,clientX:0,clientY:0,pageX:0,pageY:0,ctrlKey:0,shiftKey:0,altKey:0,metaKey:0,getModifierState:mc,button:0,buttons:0,relatedTarget:function(n){return n.relatedTarget===void 0?n.fromElement===n.srcElement?n.toElement:n.fromElement:n.relatedTarget},movementX:function(n){return"movementX"in n?n.movementX:(n!==xo&&(xo&&n.type==="mousemove"?(dc=n.screenX-xo.screenX,hc=n.screenY-xo.screenY):hc=dc=0,xo=n),dc)},movementY:function(n){return"movementY"in n?n.movementY:hc}}),zd=Un(ya),S_=L({},ya,{dataTransfer:0}),M_=Un(S_),E_=L({},vo,{relatedTarget:0}),pc=Un(E_),T_=L({},ds,{animationName:0,elapsedTime:0,pseudoElement:0}),w_=Un(T_),A_=L({},ds,{clipboardData:function(n){return"clipboardData"in n?n.clipboardData:window.clipboardData}}),C_=Un(A_),R_=L({},ds,{data:0}),Hd=Un(R_),b_={Esc:"Escape",Spacebar:" ",Left:"ArrowLeft",Up:"ArrowUp",Right:"ArrowRight",Down:"ArrowDown",Del:"Delete",Win:"OS",Menu:"ContextMenu",Apps:"ContextMenu",Scroll:"ScrollLock",MozPrintableKey:"Unidentified"},P_={8:"Backspace",9:"Tab",12:"Clear",13:"Enter",16:"Shift",17:"Control",18:"Alt",19:"Pause",20:"CapsLock",27:"Escape",32:" ",33:"PageUp",34:"PageDown",35:"End",36:"Home",37:"ArrowLeft",38:"ArrowUp",39:"ArrowRight",40:"ArrowDown",45:"Insert",46:"Delete",112:"F1",113:"F2",114:"F3",115:"F4",116:"F5",117:"F6",118:"F7",119:"F8",120:"F9",121:"F10",122:"F11",123:"F12",144:"NumLock",145:"ScrollLock",224:"Meta"},L_={Alt:"altKey",Control:"ctrlKey",Meta:"metaKey",Shift:"shiftKey"};function D_(n){var i=this.nativeEvent;return i.getModifierState?i.getModifierState(n):(n=L_[n])?!!i[n]:!1}function mc(){return D_}var N_=L({},vo,{key:function(n){if(n.key){var i=b_[n.key]||n.key;if(i!=="Unidentified")return i}return n.type==="keypress"?(n=va(n),n===13?"Enter":String.fromCharCode(n)):n.type==="keydown"||n.type==="keyup"?P_[n.keyCode]||"Unidentified":""},code:0,location:0,ctrlKey:0,shiftKey:0,altKey:0,metaKey:0,repeat:0,locale:0,getModifierState:mc,charCode:function(n){return n.type==="keypress"?va(n):0},keyCode:function(n){return n.type==="keydown"||n.type==="keyup"?n.keyCode:0},which:function(n){return n.type==="keypress"?va(n):n.type==="keydown"||n.type==="keyup"?n.keyCode:0}}),I_=Un(N_),U_=L({},ya,{pointerId:0,width:0,height:0,pressure:0,tangentialPressure:0,tiltX:0,tiltY:0,twist:0,pointerType:0,isPrimary:0}),Vd=Un(U_),F_=L({},vo,{touches:0,targetTouches:0,changedTouches:0,altKey:0,metaKey:0,ctrlKey:0,shiftKey:0,getModifierState:mc}),O_=Un(F_),k_=L({},ds,{propertyName:0,elapsedTime:0,pseudoElement:0}),B_=Un(k_),z_=L({},ya,{deltaX:function(n){return"deltaX"in n?n.deltaX:"wheelDeltaX"in n?-n.wheelDeltaX:0},deltaY:function(n){return"deltaY"in n?n.deltaY:"wheelDeltaY"in n?-n.wheelDeltaY:"wheelDelta"in n?-n.wheelDelta:0},deltaZ:0,deltaMode:0}),H_=Un(z_),V_=[9,13,27,32],gc=f&&"CompositionEvent"in window,yo=null;f&&"documentMode"in document&&(yo=document.documentMode);var G_=f&&"TextEvent"in window&&!yo,Gd=f&&(!gc||yo&&8<yo&&11>=yo),Wd=" ",jd=!1;function Xd(n,i){switch(n){case"keyup":return V_.indexOf(i.keyCode)!==-1;case"keydown":return i.keyCode!==229;case"keypress":case"mousedown":case"focusout":return!0;default:return!1}}function Yd(n){return n=n.detail,typeof n=="object"&&"data"in n?n.data:null}var hs=!1;function W_(n,i){switch(n){case"compositionend":return Yd(i);case"keypress":return i.which!==32?null:(jd=!0,Wd);case"textInput":return n=i.data,n===Wd&&jd?null:n;default:return null}}function j_(n,i){if(hs)return n==="compositionend"||!gc&&Xd(n,i)?(n=kd(),_a=uc=Ji=null,hs=!1,n):null;switch(n){case"paste":return null;case"keypress":if(!(i.ctrlKey||i.altKey||i.metaKey)||i.ctrlKey&&i.altKey){if(i.char&&1<i.char.length)return i.char;if(i.which)return String.fromCharCode(i.which)}return null;case"compositionend":return Gd&&i.locale!=="ko"?null:i.data;default:return null}}var X_={color:!0,date:!0,datetime:!0,"datetime-local":!0,email:!0,month:!0,number:!0,password:!0,range:!0,search:!0,tel:!0,text:!0,time:!0,url:!0,week:!0};function qd(n){var i=n&&n.nodeName&&n.nodeName.toLowerCase();return i==="input"?!!X_[n.type]:i==="textarea"}function $d(n,i,o,c){ft(c),i=wa(i,"onChange"),0<i.length&&(o=new fc("onChange","change",null,o,c),n.push({event:o,listeners:i}))}var So=null,Mo=null;function Y_(n){hh(n,0)}function Sa(n){var i=vs(n);if(ct(i))return n}function q_(n,i){if(n==="change")return i}var Kd=!1;if(f){var _c;if(f){var vc="oninput"in document;if(!vc){var Zd=document.createElement("div");Zd.setAttribute("oninput","return;"),vc=typeof Zd.oninput=="function"}_c=vc}else _c=!1;Kd=_c&&(!document.documentMode||9<document.documentMode)}function Qd(){So&&(So.detachEvent("onpropertychange",Jd),Mo=So=null)}function Jd(n){if(n.propertyName==="value"&&Sa(Mo)){var i=[];$d(i,Mo,n,Le(n)),Gn(Y_,i)}}function $_(n,i,o){n==="focusin"?(Qd(),So=i,Mo=o,So.attachEvent("onpropertychange",Jd)):n==="focusout"&&Qd()}function K_(n){if(n==="selectionchange"||n==="keyup"||n==="keydown")return Sa(Mo)}function Z_(n,i){if(n==="click")return Sa(i)}function Q_(n,i){if(n==="input"||n==="change")return Sa(i)}function J_(n,i){return n===i&&(n!==0||1/n===1/i)||n!==n&&i!==i}var ei=typeof Object.is=="function"?Object.is:J_;function Eo(n,i){if(ei(n,i))return!0;if(typeof n!="object"||n===null||typeof i!="object"||i===null)return!1;var o=Object.keys(n),c=Object.keys(i);if(o.length!==c.length)return!1;for(c=0;c<o.length;c++){var d=o[c];if(!h.call(i,d)||!ei(n[d],i[d]))return!1}return!0}function eh(n){for(;n&&n.firstChild;)n=n.firstChild;return n}function th(n,i){var o=eh(n);n=0;for(var c;o;){if(o.nodeType===3){if(c=n+o.textContent.length,n<=i&&c>=i)return{node:o,offset:i-n};n=c}e:{for(;o;){if(o.nextSibling){o=o.nextSibling;break e}o=o.parentNode}o=void 0}o=eh(o)}}function nh(n,i){return n&&i?n===i?!0:n&&n.nodeType===3?!1:i&&i.nodeType===3?nh(n,i.parentNode):"contains"in n?n.contains(i):n.compareDocumentPosition?!!(n.compareDocumentPosition(i)&16):!1:!1}function ih(){for(var n=window,i=z();i instanceof n.HTMLIFrameElement;){try{var o=typeof i.contentWindow.location.href=="string"}catch{o=!1}if(o)n=i.contentWindow;else break;i=z(n.document)}return i}function xc(n){var i=n&&n.nodeName&&n.nodeName.toLowerCase();return i&&(i==="input"&&(n.type==="text"||n.type==="search"||n.type==="tel"||n.type==="url"||n.type==="password")||i==="textarea"||n.contentEditable==="true")}function ev(n){var i=ih(),o=n.focusedElem,c=n.selectionRange;if(i!==o&&o&&o.ownerDocument&&nh(o.ownerDocument.documentElement,o)){if(c!==null&&xc(o)){if(i=c.start,n=c.end,n===void 0&&(n=i),"selectionStart"in o)o.selectionStart=i,o.selectionEnd=Math.min(n,o.value.length);else if(n=(i=o.ownerDocument||document)&&i.defaultView||window,n.getSelection){n=n.getSelection();var d=o.textContent.length,g=Math.min(c.start,d);c=c.end===void 0?g:Math.min(c.end,d),!n.extend&&g>c&&(d=c,c=g,g=d),d=th(o,g);var M=th(o,c);d&&M&&(n.rangeCount!==1||n.anchorNode!==d.node||n.anchorOffset!==d.offset||n.focusNode!==M.node||n.focusOffset!==M.offset)&&(i=i.createRange(),i.setStart(d.node,d.offset),n.removeAllRanges(),g>c?(n.addRange(i),n.extend(M.node,M.offset)):(i.setEnd(M.node,M.offset),n.addRange(i)))}}for(i=[],n=o;n=n.parentNode;)n.nodeType===1&&i.push({element:n,left:n.scrollLeft,top:n.scrollTop});for(typeof o.focus=="function"&&o.focus(),o=0;o<i.length;o++)n=i[o],n.element.scrollLeft=n.left,n.element.scrollTop=n.top}}var tv=f&&"documentMode"in document&&11>=document.documentMode,ps=null,yc=null,To=null,Sc=!1;function rh(n,i,o){var c=o.window===o?o.document:o.nodeType===9?o:o.ownerDocument;Sc||ps==null||ps!==z(c)||(c=ps,"selectionStart"in c&&xc(c)?c={start:c.selectionStart,end:c.selectionEnd}:(c=(c.ownerDocument&&c.ownerDocument.defaultView||window).getSelection(),c={anchorNode:c.anchorNode,anchorOffset:c.anchorOffset,focusNode:c.focusNode,focusOffset:c.focusOffset}),To&&Eo(To,c)||(To=c,c=wa(yc,"onSelect"),0<c.length&&(i=new fc("onSelect","select",null,i,o),n.push({event:i,listeners:c}),i.target=ps)))}function Ma(n,i){var o={};return o[n.toLowerCase()]=i.toLowerCase(),o["Webkit"+n]="webkit"+i,o["Moz"+n]="moz"+i,o}var ms={animationend:Ma("Animation","AnimationEnd"),animationiteration:Ma("Animation","AnimationIteration"),animationstart:Ma("Animation","AnimationStart"),transitionend:Ma("Transition","TransitionEnd")},Mc={},sh={};f&&(sh=document.createElement("div").style,"AnimationEvent"in window||(delete ms.animationend.animation,delete ms.animationiteration.animation,delete ms.animationstart.animation),"TransitionEvent"in window||delete ms.transitionend.transition);function Ea(n){if(Mc[n])return Mc[n];if(!ms[n])return n;var i=ms[n],o;for(o in i)if(i.hasOwnProperty(o)&&o in sh)return Mc[n]=i[o];return n}var oh=Ea("animationend"),ah=Ea("animationiteration"),lh=Ea("animationstart"),ch=Ea("transitionend"),uh=new Map,fh="abort auxClick cancel canPlay canPlayThrough click close contextMenu copy cut drag dragEnd dragEnter dragExit dragLeave dragOver dragStart drop durationChange emptied encrypted ended error gotPointerCapture input invalid keyDown keyPress keyUp load loadedData loadedMetadata loadStart lostPointerCapture mouseDown mouseMove mouseOut mouseOver mouseUp paste pause play playing pointerCancel pointerDown pointerMove pointerOut pointerOver pointerUp progress rateChange reset resize seeked seeking stalled submit suspend timeUpdate touchCancel touchEnd touchStart volumeChange scroll toggle touchMove waiting wheel".split(" ");function er(n,i){uh.set(n,i),l(i,[n])}for(var Ec=0;Ec<fh.length;Ec++){var Tc=fh[Ec],nv=Tc.toLowerCase(),iv=Tc[0].toUpperCase()+Tc.slice(1);er(nv,"on"+iv)}er(oh,"onAnimationEnd"),er(ah,"onAnimationIteration"),er(lh,"onAnimationStart"),er("dblclick","onDoubleClick"),er("focusin","onFocus"),er("focusout","onBlur"),er(ch,"onTransitionEnd"),u("onMouseEnter",["mouseout","mouseover"]),u("onMouseLeave",["mouseout","mouseover"]),u("onPointerEnter",["pointerout","pointerover"]),u("onPointerLeave",["pointerout","pointerover"]),l("onChange","change click focusin focusout input keydown keyup selectionchange".split(" ")),l("onSelect","focusout contextmenu dragend focusin keydown keyup mousedown mouseup selectionchange".split(" ")),l("onBeforeInput",["compositionend","keypress","textInput","paste"]),l("onCompositionEnd","compositionend focusout keydown keypress keyup mousedown".split(" ")),l("onCompositionStart","compositionstart focusout keydown keypress keyup mousedown".split(" ")),l("onCompositionUpdate","compositionupdate focusout keydown keypress keyup mousedown".split(" "));var wo="abort canplay canplaythrough durationchange emptied encrypted ended error loadeddata loadedmetadata loadstart pause play playing progress ratechange resize seeked seeking stalled suspend timeupdate volumechange waiting".split(" "),rv=new Set("cancel close invalid load scroll toggle".split(" ").concat(wo));function dh(n,i,o){var c=n.type||"unknown-event";n.currentTarget=o,ua(c,i,void 0,n),n.currentTarget=null}function hh(n,i){i=(i&4)!==0;for(var o=0;o<n.length;o++){var c=n[o],d=c.event;c=c.listeners;e:{var g=void 0;if(i)for(var M=c.length-1;0<=M;M--){var I=c[M],B=I.instance,ie=I.currentTarget;if(I=I.listener,B!==g&&d.isPropagationStopped())break e;dh(d,I,ie),g=B}else for(M=0;M<c.length;M++){if(I=c[M],B=I.instance,ie=I.currentTarget,I=I.listener,B!==g&&d.isPropagationStopped())break e;dh(d,I,ie),g=B}}}if(Mi)throw n=ls,Mi=!1,ls=null,n}function Nt(n,i){var o=i[Dc];o===void 0&&(o=i[Dc]=new Set);var c=n+"__bubble";o.has(c)||(ph(i,n,2,!1),o.add(c))}function wc(n,i,o){var c=0;i&&(c|=4),ph(o,n,c,i)}var Ta="_reactListening"+Math.random().toString(36).slice(2);function Ao(n){if(!n[Ta]){n[Ta]=!0,s.forEach(function(o){o!=="selectionchange"&&(rv.has(o)||wc(o,!1,n),wc(o,!0,n))});var i=n.nodeType===9?n:n.ownerDocument;i===null||i[Ta]||(i[Ta]=!0,wc("selectionchange",!1,i))}}function ph(n,i,o,c){switch(Od(i)){case 1:var d=v_;break;case 4:d=x_;break;default:d=lc}o=d.bind(null,i,o,n),d=void 0,!as||i!=="touchstart"&&i!=="touchmove"&&i!=="wheel"||(d=!0),c?d!==void 0?n.addEventListener(i,o,{capture:!0,passive:d}):n.addEventListener(i,o,!0):d!==void 0?n.addEventListener(i,o,{passive:d}):n.addEventListener(i,o,!1)}function Ac(n,i,o,c,d){var g=c;if((i&1)===0&&(i&2)===0&&c!==null)e:for(;;){if(c===null)return;var M=c.tag;if(M===3||M===4){var I=c.stateNode.containerInfo;if(I===d||I.nodeType===8&&I.parentNode===d)break;if(M===4)for(M=c.return;M!==null;){var B=M.tag;if((B===3||B===4)&&(B=M.stateNode.containerInfo,B===d||B.nodeType===8&&B.parentNode===d))return;M=M.return}for(;I!==null;){if(M=Lr(I),M===null)return;if(B=M.tag,B===5||B===6){c=g=M;continue e}I=I.parentNode}}c=c.return}Gn(function(){var ie=g,_e=Le(o),ye=[];e:{var ge=uh.get(n);if(ge!==void 0){var Ue=fc,He=n;switch(n){case"keypress":if(va(o)===0)break e;case"keydown":case"keyup":Ue=I_;break;case"focusin":He="focus",Ue=pc;break;case"focusout":He="blur",Ue=pc;break;case"beforeblur":case"afterblur":Ue=pc;break;case"click":if(o.button===2)break e;case"auxclick":case"dblclick":case"mousedown":case"mousemove":case"mouseup":case"mouseout":case"mouseover":case"contextmenu":Ue=zd;break;case"drag":case"dragend":case"dragenter":case"dragexit":case"dragleave":case"dragover":case"dragstart":case"drop":Ue=M_;break;case"touchcancel":case"touchend":case"touchmove":case"touchstart":Ue=O_;break;case oh:case ah:case lh:Ue=w_;break;case ch:Ue=B_;break;case"scroll":Ue=y_;break;case"wheel":Ue=H_;break;case"copy":case"cut":case"paste":Ue=C_;break;case"gotpointercapture":case"lostpointercapture":case"pointercancel":case"pointerdown":case"pointermove":case"pointerout":case"pointerover":case"pointerup":Ue=Vd}var Ve=(i&4)!==0,Wt=!Ve&&n==="scroll",q=Ve?ge!==null?ge+"Capture":null:ge;Ve=[];for(var H=ie,$;H!==null;){$=H;var Te=$.stateNode;if($.tag===5&&Te!==null&&($=Te,q!==null&&(Te=Xi(H,q),Te!=null&&Ve.push(Co(H,Te,$)))),Wt)break;H=H.return}0<Ve.length&&(ge=new Ue(ge,He,null,o,_e),ye.push({event:ge,listeners:Ve}))}}if((i&7)===0){e:{if(ge=n==="mouseover"||n==="pointerover",Ue=n==="mouseout"||n==="pointerout",ge&&o!==G&&(He=o.relatedTarget||o.fromElement)&&(Lr(He)||He[wi]))break e;if((Ue||ge)&&(ge=_e.window===_e?_e:(ge=_e.ownerDocument)?ge.defaultView||ge.parentWindow:window,Ue?(He=o.relatedTarget||o.toElement,Ue=ie,He=He?Lr(He):null,He!==null&&(Wt=Ei(He),He!==Wt||He.tag!==5&&He.tag!==6)&&(He=null)):(Ue=null,He=ie),Ue!==He)){if(Ve=zd,Te="onMouseLeave",q="onMouseEnter",H="mouse",(n==="pointerout"||n==="pointerover")&&(Ve=Vd,Te="onPointerLeave",q="onPointerEnter",H="pointer"),Wt=Ue==null?ge:vs(Ue),$=He==null?ge:vs(He),ge=new Ve(Te,H+"leave",Ue,o,_e),ge.target=Wt,ge.relatedTarget=$,Te=null,Lr(_e)===ie&&(Ve=new Ve(q,H+"enter",He,o,_e),Ve.target=$,Ve.relatedTarget=Wt,Te=Ve),Wt=Te,Ue&&He)t:{for(Ve=Ue,q=He,H=0,$=Ve;$;$=gs($))H++;for($=0,Te=q;Te;Te=gs(Te))$++;for(;0<H-$;)Ve=gs(Ve),H--;for(;0<$-H;)q=gs(q),$--;for(;H--;){if(Ve===q||q!==null&&Ve===q.alternate)break t;Ve=gs(Ve),q=gs(q)}Ve=null}else Ve=null;Ue!==null&&mh(ye,ge,Ue,Ve,!1),He!==null&&Wt!==null&&mh(ye,Wt,He,Ve,!0)}}e:{if(ge=ie?vs(ie):window,Ue=ge.nodeName&&ge.nodeName.toLowerCase(),Ue==="select"||Ue==="input"&&ge.type==="file")var Xe=q_;else if(qd(ge))if(Kd)Xe=Q_;else{Xe=K_;var qe=$_}else(Ue=ge.nodeName)&&Ue.toLowerCase()==="input"&&(ge.type==="checkbox"||ge.type==="radio")&&(Xe=Z_);if(Xe&&(Xe=Xe(n,ie))){$d(ye,Xe,o,_e);break e}qe&&qe(n,ge,ie),n==="focusout"&&(qe=ge._wrapperState)&&qe.controlled&&ge.type==="number"&&Je(ge,"number",ge.value)}switch(qe=ie?vs(ie):window,n){case"focusin":(qd(qe)||qe.contentEditable==="true")&&(ps=qe,yc=ie,To=null);break;case"focusout":To=yc=ps=null;break;case"mousedown":Sc=!0;break;case"contextmenu":case"mouseup":case"dragend":Sc=!1,rh(ye,o,_e);break;case"selectionchange":if(tv)break;case"keydown":case"keyup":rh(ye,o,_e)}var $e;if(gc)e:{switch(n){case"compositionstart":var nt="onCompositionStart";break e;case"compositionend":nt="onCompositionEnd";break e;case"compositionupdate":nt="onCompositionUpdate";break e}nt=void 0}else hs?Xd(n,o)&&(nt="onCompositionEnd"):n==="keydown"&&o.keyCode===229&&(nt="onCompositionStart");nt&&(Gd&&o.locale!=="ko"&&(hs||nt!=="onCompositionStart"?nt==="onCompositionEnd"&&hs&&($e=kd()):(Ji=_e,uc="value"in Ji?Ji.value:Ji.textContent,hs=!0)),qe=wa(ie,nt),0<qe.length&&(nt=new Hd(nt,n,null,o,_e),ye.push({event:nt,listeners:qe}),$e?nt.data=$e:($e=Yd(o),$e!==null&&(nt.data=$e)))),($e=G_?W_(n,o):j_(n,o))&&(ie=wa(ie,"onBeforeInput"),0<ie.length&&(_e=new Hd("onBeforeInput","beforeinput",null,o,_e),ye.push({event:_e,listeners:ie}),_e.data=$e))}hh(ye,i)})}function Co(n,i,o){return{instance:n,listener:i,currentTarget:o}}function wa(n,i){for(var o=i+"Capture",c=[];n!==null;){var d=n,g=d.stateNode;d.tag===5&&g!==null&&(d=g,g=Xi(n,o),g!=null&&c.unshift(Co(n,g,d)),g=Xi(n,i),g!=null&&c.push(Co(n,g,d))),n=n.return}return c}function gs(n){if(n===null)return null;do n=n.return;while(n&&n.tag!==5);return n||null}function mh(n,i,o,c,d){for(var g=i._reactName,M=[];o!==null&&o!==c;){var I=o,B=I.alternate,ie=I.stateNode;if(B!==null&&B===c)break;I.tag===5&&ie!==null&&(I=ie,d?(B=Xi(o,g),B!=null&&M.unshift(Co(o,B,I))):d||(B=Xi(o,g),B!=null&&M.push(Co(o,B,I)))),o=o.return}M.length!==0&&n.push({event:i,listeners:M})}var sv=/\r\n?/g,ov=/\u0000|\uFFFD/g;function gh(n){return(typeof n=="string"?n:""+n).replace(sv,`
`).replace(ov,"")}function Aa(n,i,o){if(i=gh(i),gh(n)!==i&&o)throw Error(t(425))}function Ca(){}var Cc=null,Rc=null;function bc(n,i){return n==="textarea"||n==="noscript"||typeof i.children=="string"||typeof i.children=="number"||typeof i.dangerouslySetInnerHTML=="object"&&i.dangerouslySetInnerHTML!==null&&i.dangerouslySetInnerHTML.__html!=null}var Pc=typeof setTimeout=="function"?setTimeout:void 0,av=typeof clearTimeout=="function"?clearTimeout:void 0,_h=typeof Promise=="function"?Promise:void 0,lv=typeof queueMicrotask=="function"?queueMicrotask:typeof _h<"u"?function(n){return _h.resolve(null).then(n).catch(cv)}:Pc;function cv(n){setTimeout(function(){throw n})}function Lc(n,i){var o=i,c=0;do{var d=o.nextSibling;if(n.removeChild(o),d&&d.nodeType===8)if(o=d.data,o==="/$"){if(c===0){n.removeChild(d),_o(i);return}c--}else o!=="$"&&o!=="$?"&&o!=="$!"||c++;o=d}while(o);_o(i)}function tr(n){for(;n!=null;n=n.nextSibling){var i=n.nodeType;if(i===1||i===3)break;if(i===8){if(i=n.data,i==="$"||i==="$!"||i==="$?")break;if(i==="/$")return null}}return n}function vh(n){n=n.previousSibling;for(var i=0;n;){if(n.nodeType===8){var o=n.data;if(o==="$"||o==="$!"||o==="$?"){if(i===0)return n;i--}else o==="/$"&&i++}n=n.previousSibling}return null}var _s=Math.random().toString(36).slice(2),pi="__reactFiber$"+_s,Ro="__reactProps$"+_s,wi="__reactContainer$"+_s,Dc="__reactEvents$"+_s,uv="__reactListeners$"+_s,fv="__reactHandles$"+_s;function Lr(n){var i=n[pi];if(i)return i;for(var o=n.parentNode;o;){if(i=o[wi]||o[pi]){if(o=i.alternate,i.child!==null||o!==null&&o.child!==null)for(n=vh(n);n!==null;){if(o=n[pi])return o;n=vh(n)}return i}n=o,o=n.parentNode}return null}function bo(n){return n=n[pi]||n[wi],!n||n.tag!==5&&n.tag!==6&&n.tag!==13&&n.tag!==3?null:n}function vs(n){if(n.tag===5||n.tag===6)return n.stateNode;throw Error(t(33))}function Ra(n){return n[Ro]||null}var Nc=[],xs=-1;function nr(n){return{current:n}}function It(n){0>xs||(n.current=Nc[xs],Nc[xs]=null,xs--)}function Dt(n,i){xs++,Nc[xs]=n.current,n.current=i}var ir={},dn=nr(ir),En=nr(!1),Dr=ir;function ys(n,i){var o=n.type.contextTypes;if(!o)return ir;var c=n.stateNode;if(c&&c.__reactInternalMemoizedUnmaskedChildContext===i)return c.__reactInternalMemoizedMaskedChildContext;var d={},g;for(g in o)d[g]=i[g];return c&&(n=n.stateNode,n.__reactInternalMemoizedUnmaskedChildContext=i,n.__reactInternalMemoizedMaskedChildContext=d),d}function Tn(n){return n=n.childContextTypes,n!=null}function ba(){It(En),It(dn)}function xh(n,i,o){if(dn.current!==ir)throw Error(t(168));Dt(dn,i),Dt(En,o)}function yh(n,i,o){var c=n.stateNode;if(i=i.childContextTypes,typeof c.getChildContext!="function")return o;c=c.getChildContext();for(var d in c)if(!(d in i))throw Error(t(108,Me(n)||"Unknown",d));return L({},o,c)}function Pa(n){return n=(n=n.stateNode)&&n.__reactInternalMemoizedMergedChildContext||ir,Dr=dn.current,Dt(dn,n),Dt(En,En.current),!0}function Sh(n,i,o){var c=n.stateNode;if(!c)throw Error(t(169));o?(n=yh(n,i,Dr),c.__reactInternalMemoizedMergedChildContext=n,It(En),It(dn),Dt(dn,n)):It(En),Dt(En,o)}var Ai=null,La=!1,Ic=!1;function Mh(n){Ai===null?Ai=[n]:Ai.push(n)}function dv(n){La=!0,Mh(n)}function rr(){if(!Ic&&Ai!==null){Ic=!0;var n=0,i=wt;try{var o=Ai;for(wt=1;n<o.length;n++){var c=o[n];do c=c(!0);while(c!==null)}Ai=null,La=!1}catch(d){throw Ai!==null&&(Ai=Ai.slice(n+1)),ne(Ke,rr),d}finally{wt=i,Ic=!1}}return null}var Ss=[],Ms=0,Da=null,Na=0,Wn=[],jn=0,Nr=null,Ci=1,Ri="";function Ir(n,i){Ss[Ms++]=Na,Ss[Ms++]=Da,Da=n,Na=i}function Eh(n,i,o){Wn[jn++]=Ci,Wn[jn++]=Ri,Wn[jn++]=Nr,Nr=n;var c=Ci;n=Ri;var d=32-ke(c)-1;c&=~(1<<d),o+=1;var g=32-ke(i)+d;if(30<g){var M=d-d%5;g=(c&(1<<M)-1).toString(32),c>>=M,d-=M,Ci=1<<32-ke(i)+d|o<<d|c,Ri=g+n}else Ci=1<<g|o<<d|c,Ri=n}function Uc(n){n.return!==null&&(Ir(n,1),Eh(n,1,0))}function Fc(n){for(;n===Da;)Da=Ss[--Ms],Ss[Ms]=null,Na=Ss[--Ms],Ss[Ms]=null;for(;n===Nr;)Nr=Wn[--jn],Wn[jn]=null,Ri=Wn[--jn],Wn[jn]=null,Ci=Wn[--jn],Wn[jn]=null}var Fn=null,On=null,Ot=!1,ti=null;function Th(n,i){var o=$n(5,null,null,0);o.elementType="DELETED",o.stateNode=i,o.return=n,i=n.deletions,i===null?(n.deletions=[o],n.flags|=16):i.push(o)}function wh(n,i){switch(n.tag){case 5:var o=n.type;return i=i.nodeType!==1||o.toLowerCase()!==i.nodeName.toLowerCase()?null:i,i!==null?(n.stateNode=i,Fn=n,On=tr(i.firstChild),!0):!1;case 6:return i=n.pendingProps===""||i.nodeType!==3?null:i,i!==null?(n.stateNode=i,Fn=n,On=null,!0):!1;case 13:return i=i.nodeType!==8?null:i,i!==null?(o=Nr!==null?{id:Ci,overflow:Ri}:null,n.memoizedState={dehydrated:i,treeContext:o,retryLane:1073741824},o=$n(18,null,null,0),o.stateNode=i,o.return=n,n.child=o,Fn=n,On=null,!0):!1;default:return!1}}function Oc(n){return(n.mode&1)!==0&&(n.flags&128)===0}function kc(n){if(Ot){var i=On;if(i){var o=i;if(!wh(n,i)){if(Oc(n))throw Error(t(418));i=tr(o.nextSibling);var c=Fn;i&&wh(n,i)?Th(c,o):(n.flags=n.flags&-4097|2,Ot=!1,Fn=n)}}else{if(Oc(n))throw Error(t(418));n.flags=n.flags&-4097|2,Ot=!1,Fn=n}}}function Ah(n){for(n=n.return;n!==null&&n.tag!==5&&n.tag!==3&&n.tag!==13;)n=n.return;Fn=n}function Ia(n){if(n!==Fn)return!1;if(!Ot)return Ah(n),Ot=!0,!1;var i;if((i=n.tag!==3)&&!(i=n.tag!==5)&&(i=n.type,i=i!=="head"&&i!=="body"&&!bc(n.type,n.memoizedProps)),i&&(i=On)){if(Oc(n))throw Ch(),Error(t(418));for(;i;)Th(n,i),i=tr(i.nextSibling)}if(Ah(n),n.tag===13){if(n=n.memoizedState,n=n!==null?n.dehydrated:null,!n)throw Error(t(317));e:{for(n=n.nextSibling,i=0;n;){if(n.nodeType===8){var o=n.data;if(o==="/$"){if(i===0){On=tr(n.nextSibling);break e}i--}else o!=="$"&&o!=="$!"&&o!=="$?"||i++}n=n.nextSibling}On=null}}else On=Fn?tr(n.stateNode.nextSibling):null;return!0}function Ch(){for(var n=On;n;)n=tr(n.nextSibling)}function Es(){On=Fn=null,Ot=!1}function Bc(n){ti===null?ti=[n]:ti.push(n)}var hv=D.ReactCurrentBatchConfig;function ni(n,i){if(n&&n.defaultProps){i=L({},i),n=n.defaultProps;for(var o in n)i[o]===void 0&&(i[o]=n[o]);return i}return i}var Ua=nr(null),Fa=null,Ts=null,zc=null;function Hc(){zc=Ts=Fa=null}function Vc(n){var i=Ua.current;It(Ua),n._currentValue=i}function Gc(n,i,o){for(;n!==null;){var c=n.alternate;if((n.childLanes&i)!==i?(n.childLanes|=i,c!==null&&(c.childLanes|=i)):c!==null&&(c.childLanes&i)!==i&&(c.childLanes|=i),n===o)break;n=n.return}}function ws(n,i){Fa=n,zc=Ts=null,n=n.dependencies,n!==null&&n.firstContext!==null&&((n.lanes&i)!==0&&(wn=!0),n.firstContext=null)}function Xn(n){var i=n._currentValue;if(zc!==n)if(n={context:n,memoizedValue:i,next:null},Ts===null){if(Fa===null)throw Error(t(308));Ts=n,Fa.dependencies={lanes:0,firstContext:n}}else Ts=Ts.next=n;return i}var Ur=null;function Wc(n){Ur===null?Ur=[n]:Ur.push(n)}function Rh(n,i,o,c){var d=i.interleaved;return d===null?(o.next=o,Wc(i)):(o.next=d.next,d.next=o),i.interleaved=o,bi(n,c)}function bi(n,i){n.lanes|=i;var o=n.alternate;for(o!==null&&(o.lanes|=i),o=n,n=n.return;n!==null;)n.childLanes|=i,o=n.alternate,o!==null&&(o.childLanes|=i),o=n,n=n.return;return o.tag===3?o.stateNode:null}var sr=!1;function jc(n){n.updateQueue={baseState:n.memoizedState,firstBaseUpdate:null,lastBaseUpdate:null,shared:{pending:null,interleaved:null,lanes:0},effects:null}}function bh(n,i){n=n.updateQueue,i.updateQueue===n&&(i.updateQueue={baseState:n.baseState,firstBaseUpdate:n.firstBaseUpdate,lastBaseUpdate:n.lastBaseUpdate,shared:n.shared,effects:n.effects})}function Pi(n,i){return{eventTime:n,lane:i,tag:0,payload:null,callback:null,next:null}}function or(n,i,o){var c=n.updateQueue;if(c===null)return null;if(c=c.shared,(_t&2)!==0){var d=c.pending;return d===null?i.next=i:(i.next=d.next,d.next=i),c.pending=i,bi(n,o)}return d=c.interleaved,d===null?(i.next=i,Wc(c)):(i.next=d.next,d.next=i),c.interleaved=i,bi(n,o)}function Oa(n,i,o){if(i=i.updateQueue,i!==null&&(i=i.shared,(o&4194240)!==0)){var c=i.lanes;c&=n.pendingLanes,o|=c,i.lanes=o,sc(n,o)}}function Ph(n,i){var o=n.updateQueue,c=n.alternate;if(c!==null&&(c=c.updateQueue,o===c)){var d=null,g=null;if(o=o.firstBaseUpdate,o!==null){do{var M={eventTime:o.eventTime,lane:o.lane,tag:o.tag,payload:o.payload,callback:o.callback,next:null};g===null?d=g=M:g=g.next=M,o=o.next}while(o!==null);g===null?d=g=i:g=g.next=i}else d=g=i;o={baseState:c.baseState,firstBaseUpdate:d,lastBaseUpdate:g,shared:c.shared,effects:c.effects},n.updateQueue=o;return}n=o.lastBaseUpdate,n===null?o.firstBaseUpdate=i:n.next=i,o.lastBaseUpdate=i}function ka(n,i,o,c){var d=n.updateQueue;sr=!1;var g=d.firstBaseUpdate,M=d.lastBaseUpdate,I=d.shared.pending;if(I!==null){d.shared.pending=null;var B=I,ie=B.next;B.next=null,M===null?g=ie:M.next=ie,M=B;var _e=n.alternate;_e!==null&&(_e=_e.updateQueue,I=_e.lastBaseUpdate,I!==M&&(I===null?_e.firstBaseUpdate=ie:I.next=ie,_e.lastBaseUpdate=B))}if(g!==null){var ye=d.baseState;M=0,_e=ie=B=null,I=g;do{var ge=I.lane,Ue=I.eventTime;if((c&ge)===ge){_e!==null&&(_e=_e.next={eventTime:Ue,lane:0,tag:I.tag,payload:I.payload,callback:I.callback,next:null});e:{var He=n,Ve=I;switch(ge=i,Ue=o,Ve.tag){case 1:if(He=Ve.payload,typeof He=="function"){ye=He.call(Ue,ye,ge);break e}ye=He;break e;case 3:He.flags=He.flags&-65537|128;case 0:if(He=Ve.payload,ge=typeof He=="function"?He.call(Ue,ye,ge):He,ge==null)break e;ye=L({},ye,ge);break e;case 2:sr=!0}}I.callback!==null&&I.lane!==0&&(n.flags|=64,ge=d.effects,ge===null?d.effects=[I]:ge.push(I))}else Ue={eventTime:Ue,lane:ge,tag:I.tag,payload:I.payload,callback:I.callback,next:null},_e===null?(ie=_e=Ue,B=ye):_e=_e.next=Ue,M|=ge;if(I=I.next,I===null){if(I=d.shared.pending,I===null)break;ge=I,I=ge.next,ge.next=null,d.lastBaseUpdate=ge,d.shared.pending=null}}while(!0);if(_e===null&&(B=ye),d.baseState=B,d.firstBaseUpdate=ie,d.lastBaseUpdate=_e,i=d.shared.interleaved,i!==null){d=i;do M|=d.lane,d=d.next;while(d!==i)}else g===null&&(d.shared.lanes=0);kr|=M,n.lanes=M,n.memoizedState=ye}}function Lh(n,i,o){if(n=i.effects,i.effects=null,n!==null)for(i=0;i<n.length;i++){var c=n[i],d=c.callback;if(d!==null){if(c.callback=null,c=o,typeof d!="function")throw Error(t(191,d));d.call(c)}}}var Dh=new r.Component().refs;function Xc(n,i,o,c){i=n.memoizedState,o=o(c,i),o=o==null?i:L({},i,o),n.memoizedState=o,n.lanes===0&&(n.updateQueue.baseState=o)}var Ba={isMounted:function(n){return(n=n._reactInternals)?Ei(n)===n:!1},enqueueSetState:function(n,i,o){n=n._reactInternals;var c=yn(),d=ur(n),g=Pi(c,d);g.payload=i,o!=null&&(g.callback=o),i=or(n,g,d),i!==null&&(si(i,n,d,c),Oa(i,n,d))},enqueueReplaceState:function(n,i,o){n=n._reactInternals;var c=yn(),d=ur(n),g=Pi(c,d);g.tag=1,g.payload=i,o!=null&&(g.callback=o),i=or(n,g,d),i!==null&&(si(i,n,d,c),Oa(i,n,d))},enqueueForceUpdate:function(n,i){n=n._reactInternals;var o=yn(),c=ur(n),d=Pi(o,c);d.tag=2,i!=null&&(d.callback=i),i=or(n,d,c),i!==null&&(si(i,n,c,o),Oa(i,n,c))}};function Nh(n,i,o,c,d,g,M){return n=n.stateNode,typeof n.shouldComponentUpdate=="function"?n.shouldComponentUpdate(c,g,M):i.prototype&&i.prototype.isPureReactComponent?!Eo(o,c)||!Eo(d,g):!0}function Ih(n,i,o){var c=!1,d=ir,g=i.contextType;return typeof g=="object"&&g!==null?g=Xn(g):(d=Tn(i)?Dr:dn.current,c=i.contextTypes,g=(c=c!=null)?ys(n,d):ir),i=new i(o,g),n.memoizedState=i.state!==null&&i.state!==void 0?i.state:null,i.updater=Ba,n.stateNode=i,i._reactInternals=n,c&&(n=n.stateNode,n.__reactInternalMemoizedUnmaskedChildContext=d,n.__reactInternalMemoizedMaskedChildContext=g),i}function Uh(n,i,o,c){n=i.state,typeof i.componentWillReceiveProps=="function"&&i.componentWillReceiveProps(o,c),typeof i.UNSAFE_componentWillReceiveProps=="function"&&i.UNSAFE_componentWillReceiveProps(o,c),i.state!==n&&Ba.enqueueReplaceState(i,i.state,null)}function Yc(n,i,o,c){var d=n.stateNode;d.props=o,d.state=n.memoizedState,d.refs=Dh,jc(n);var g=i.contextType;typeof g=="object"&&g!==null?d.context=Xn(g):(g=Tn(i)?Dr:dn.current,d.context=ys(n,g)),d.state=n.memoizedState,g=i.getDerivedStateFromProps,typeof g=="function"&&(Xc(n,i,g,o),d.state=n.memoizedState),typeof i.getDerivedStateFromProps=="function"||typeof d.getSnapshotBeforeUpdate=="function"||typeof d.UNSAFE_componentWillMount!="function"&&typeof d.componentWillMount!="function"||(i=d.state,typeof d.componentWillMount=="function"&&d.componentWillMount(),typeof d.UNSAFE_componentWillMount=="function"&&d.UNSAFE_componentWillMount(),i!==d.state&&Ba.enqueueReplaceState(d,d.state,null),ka(n,o,d,c),d.state=n.memoizedState),typeof d.componentDidMount=="function"&&(n.flags|=4194308)}function Po(n,i,o){if(n=o.ref,n!==null&&typeof n!="function"&&typeof n!="object"){if(o._owner){if(o=o._owner,o){if(o.tag!==1)throw Error(t(309));var c=o.stateNode}if(!c)throw Error(t(147,n));var d=c,g=""+n;return i!==null&&i.ref!==null&&typeof i.ref=="function"&&i.ref._stringRef===g?i.ref:(i=function(M){var I=d.refs;I===Dh&&(I=d.refs={}),M===null?delete I[g]:I[g]=M},i._stringRef=g,i)}if(typeof n!="string")throw Error(t(284));if(!o._owner)throw Error(t(290,n))}return n}function za(n,i){throw n=Object.prototype.toString.call(i),Error(t(31,n==="[object Object]"?"object with keys {"+Object.keys(i).join(", ")+"}":n))}function Fh(n){var i=n._init;return i(n._payload)}function Oh(n){function i(q,H){if(n){var $=q.deletions;$===null?(q.deletions=[H],q.flags|=16):$.push(H)}}function o(q,H){if(!n)return null;for(;H!==null;)i(q,H),H=H.sibling;return null}function c(q,H){for(q=new Map;H!==null;)H.key!==null?q.set(H.key,H):q.set(H.index,H),H=H.sibling;return q}function d(q,H){return q=dr(q,H),q.index=0,q.sibling=null,q}function g(q,H,$){return q.index=$,n?($=q.alternate,$!==null?($=$.index,$<H?(q.flags|=2,H):$):(q.flags|=2,H)):(q.flags|=1048576,H)}function M(q){return n&&q.alternate===null&&(q.flags|=2),q}function I(q,H,$,Te){return H===null||H.tag!==6?(H=Pu($,q.mode,Te),H.return=q,H):(H=d(H,$),H.return=q,H)}function B(q,H,$,Te){var Xe=$.type;return Xe===U?_e(q,H,$.props.children,Te,$.key):H!==null&&(H.elementType===Xe||typeof Xe=="object"&&Xe!==null&&Xe.$$typeof===Z&&Fh(Xe)===H.type)?(Te=d(H,$.props),Te.ref=Po(q,H,$),Te.return=q,Te):(Te=sl($.type,$.key,$.props,null,q.mode,Te),Te.ref=Po(q,H,$),Te.return=q,Te)}function ie(q,H,$,Te){return H===null||H.tag!==4||H.stateNode.containerInfo!==$.containerInfo||H.stateNode.implementation!==$.implementation?(H=Lu($,q.mode,Te),H.return=q,H):(H=d(H,$.children||[]),H.return=q,H)}function _e(q,H,$,Te,Xe){return H===null||H.tag!==7?(H=Vr($,q.mode,Te,Xe),H.return=q,H):(H=d(H,$),H.return=q,H)}function ye(q,H,$){if(typeof H=="string"&&H!==""||typeof H=="number")return H=Pu(""+H,q.mode,$),H.return=q,H;if(typeof H=="object"&&H!==null){switch(H.$$typeof){case V:return $=sl(H.type,H.key,H.props,null,q.mode,$),$.ref=Po(q,null,H),$.return=q,$;case O:return H=Lu(H,q.mode,$),H.return=q,H;case Z:var Te=H._init;return ye(q,Te(H._payload),$)}if(N(H)||se(H))return H=Vr(H,q.mode,$,null),H.return=q,H;za(q,H)}return null}function ge(q,H,$,Te){var Xe=H!==null?H.key:null;if(typeof $=="string"&&$!==""||typeof $=="number")return Xe!==null?null:I(q,H,""+$,Te);if(typeof $=="object"&&$!==null){switch($.$$typeof){case V:return $.key===Xe?B(q,H,$,Te):null;case O:return $.key===Xe?ie(q,H,$,Te):null;case Z:return Xe=$._init,ge(q,H,Xe($._payload),Te)}if(N($)||se($))return Xe!==null?null:_e(q,H,$,Te,null);za(q,$)}return null}function Ue(q,H,$,Te,Xe){if(typeof Te=="string"&&Te!==""||typeof Te=="number")return q=q.get($)||null,I(H,q,""+Te,Xe);if(typeof Te=="object"&&Te!==null){switch(Te.$$typeof){case V:return q=q.get(Te.key===null?$:Te.key)||null,B(H,q,Te,Xe);case O:return q=q.get(Te.key===null?$:Te.key)||null,ie(H,q,Te,Xe);case Z:var qe=Te._init;return Ue(q,H,$,qe(Te._payload),Xe)}if(N(Te)||se(Te))return q=q.get($)||null,_e(H,q,Te,Xe,null);za(H,Te)}return null}function He(q,H,$,Te){for(var Xe=null,qe=null,$e=H,nt=H=0,sn=null;$e!==null&&nt<$.length;nt++){$e.index>nt?(sn=$e,$e=null):sn=$e.sibling;var St=ge(q,$e,$[nt],Te);if(St===null){$e===null&&($e=sn);break}n&&$e&&St.alternate===null&&i(q,$e),H=g(St,H,nt),qe===null?Xe=St:qe.sibling=St,qe=St,$e=sn}if(nt===$.length)return o(q,$e),Ot&&Ir(q,nt),Xe;if($e===null){for(;nt<$.length;nt++)$e=ye(q,$[nt],Te),$e!==null&&(H=g($e,H,nt),qe===null?Xe=$e:qe.sibling=$e,qe=$e);return Ot&&Ir(q,nt),Xe}for($e=c(q,$e);nt<$.length;nt++)sn=Ue($e,q,nt,$[nt],Te),sn!==null&&(n&&sn.alternate!==null&&$e.delete(sn.key===null?nt:sn.key),H=g(sn,H,nt),qe===null?Xe=sn:qe.sibling=sn,qe=sn);return n&&$e.forEach(function(hr){return i(q,hr)}),Ot&&Ir(q,nt),Xe}function Ve(q,H,$,Te){var Xe=se($);if(typeof Xe!="function")throw Error(t(150));if($=Xe.call($),$==null)throw Error(t(151));for(var qe=Xe=null,$e=H,nt=H=0,sn=null,St=$.next();$e!==null&&!St.done;nt++,St=$.next()){$e.index>nt?(sn=$e,$e=null):sn=$e.sibling;var hr=ge(q,$e,St.value,Te);if(hr===null){$e===null&&($e=sn);break}n&&$e&&hr.alternate===null&&i(q,$e),H=g(hr,H,nt),qe===null?Xe=hr:qe.sibling=hr,qe=hr,$e=sn}if(St.done)return o(q,$e),Ot&&Ir(q,nt),Xe;if($e===null){for(;!St.done;nt++,St=$.next())St=ye(q,St.value,Te),St!==null&&(H=g(St,H,nt),qe===null?Xe=St:qe.sibling=St,qe=St);return Ot&&Ir(q,nt),Xe}for($e=c(q,$e);!St.done;nt++,St=$.next())St=Ue($e,q,nt,St.value,Te),St!==null&&(n&&St.alternate!==null&&$e.delete(St.key===null?nt:St.key),H=g(St,H,nt),qe===null?Xe=St:qe.sibling=St,qe=St);return n&&$e.forEach(function(Xv){return i(q,Xv)}),Ot&&Ir(q,nt),Xe}function Wt(q,H,$,Te){if(typeof $=="object"&&$!==null&&$.type===U&&$.key===null&&($=$.props.children),typeof $=="object"&&$!==null){switch($.$$typeof){case V:e:{for(var Xe=$.key,qe=H;qe!==null;){if(qe.key===Xe){if(Xe=$.type,Xe===U){if(qe.tag===7){o(q,qe.sibling),H=d(qe,$.props.children),H.return=q,q=H;break e}}else if(qe.elementType===Xe||typeof Xe=="object"&&Xe!==null&&Xe.$$typeof===Z&&Fh(Xe)===qe.type){o(q,qe.sibling),H=d(qe,$.props),H.ref=Po(q,qe,$),H.return=q,q=H;break e}o(q,qe);break}else i(q,qe);qe=qe.sibling}$.type===U?(H=Vr($.props.children,q.mode,Te,$.key),H.return=q,q=H):(Te=sl($.type,$.key,$.props,null,q.mode,Te),Te.ref=Po(q,H,$),Te.return=q,q=Te)}return M(q);case O:e:{for(qe=$.key;H!==null;){if(H.key===qe)if(H.tag===4&&H.stateNode.containerInfo===$.containerInfo&&H.stateNode.implementation===$.implementation){o(q,H.sibling),H=d(H,$.children||[]),H.return=q,q=H;break e}else{o(q,H);break}else i(q,H);H=H.sibling}H=Lu($,q.mode,Te),H.return=q,q=H}return M(q);case Z:return qe=$._init,Wt(q,H,qe($._payload),Te)}if(N($))return He(q,H,$,Te);if(se($))return Ve(q,H,$,Te);za(q,$)}return typeof $=="string"&&$!==""||typeof $=="number"?($=""+$,H!==null&&H.tag===6?(o(q,H.sibling),H=d(H,$),H.return=q,q=H):(o(q,H),H=Pu($,q.mode,Te),H.return=q,q=H),M(q)):o(q,H)}return Wt}var As=Oh(!0),kh=Oh(!1),Lo={},mi=nr(Lo),Do=nr(Lo),No=nr(Lo);function Fr(n){if(n===Lo)throw Error(t(174));return n}function qc(n,i){switch(Dt(No,i),Dt(Do,n),Dt(mi,Lo),n=i.nodeType,n){case 9:case 11:i=(i=i.documentElement)?i.namespaceURI:Ce(null,"");break;default:n=n===8?i.parentNode:i,i=n.namespaceURI||null,n=n.tagName,i=Ce(i,n)}It(mi),Dt(mi,i)}function Cs(){It(mi),It(Do),It(No)}function Bh(n){Fr(No.current);var i=Fr(mi.current),o=Ce(i,n.type);i!==o&&(Dt(Do,n),Dt(mi,o))}function $c(n){Do.current===n&&(It(mi),It(Do))}var Bt=nr(0);function Ha(n){for(var i=n;i!==null;){if(i.tag===13){var o=i.memoizedState;if(o!==null&&(o=o.dehydrated,o===null||o.data==="$?"||o.data==="$!"))return i}else if(i.tag===19&&i.memoizedProps.revealOrder!==void 0){if((i.flags&128)!==0)return i}else if(i.child!==null){i.child.return=i,i=i.child;continue}if(i===n)break;for(;i.sibling===null;){if(i.return===null||i.return===n)return null;i=i.return}i.sibling.return=i.return,i=i.sibling}return null}var Kc=[];function Zc(){for(var n=0;n<Kc.length;n++)Kc[n]._workInProgressVersionPrimary=null;Kc.length=0}var Va=D.ReactCurrentDispatcher,Qc=D.ReactCurrentBatchConfig,Or=0,zt=null,$t=null,nn=null,Ga=!1,Io=!1,Uo=0,pv=0;function hn(){throw Error(t(321))}function Jc(n,i){if(i===null)return!1;for(var o=0;o<i.length&&o<n.length;o++)if(!ei(n[o],i[o]))return!1;return!0}function eu(n,i,o,c,d,g){if(Or=g,zt=i,i.memoizedState=null,i.updateQueue=null,i.lanes=0,Va.current=n===null||n.memoizedState===null?vv:xv,n=o(c,d),Io){g=0;do{if(Io=!1,Uo=0,25<=g)throw Error(t(301));g+=1,nn=$t=null,i.updateQueue=null,Va.current=yv,n=o(c,d)}while(Io)}if(Va.current=Xa,i=$t!==null&&$t.next!==null,Or=0,nn=$t=zt=null,Ga=!1,i)throw Error(t(300));return n}function tu(){var n=Uo!==0;return Uo=0,n}function gi(){var n={memoizedState:null,baseState:null,baseQueue:null,queue:null,next:null};return nn===null?zt.memoizedState=nn=n:nn=nn.next=n,nn}function Yn(){if($t===null){var n=zt.alternate;n=n!==null?n.memoizedState:null}else n=$t.next;var i=nn===null?zt.memoizedState:nn.next;if(i!==null)nn=i,$t=n;else{if(n===null)throw Error(t(310));$t=n,n={memoizedState:$t.memoizedState,baseState:$t.baseState,baseQueue:$t.baseQueue,queue:$t.queue,next:null},nn===null?zt.memoizedState=nn=n:nn=nn.next=n}return nn}function Fo(n,i){return typeof i=="function"?i(n):i}function nu(n){var i=Yn(),o=i.queue;if(o===null)throw Error(t(311));o.lastRenderedReducer=n;var c=$t,d=c.baseQueue,g=o.pending;if(g!==null){if(d!==null){var M=d.next;d.next=g.next,g.next=M}c.baseQueue=d=g,o.pending=null}if(d!==null){g=d.next,c=c.baseState;var I=M=null,B=null,ie=g;do{var _e=ie.lane;if((Or&_e)===_e)B!==null&&(B=B.next={lane:0,action:ie.action,hasEagerState:ie.hasEagerState,eagerState:ie.eagerState,next:null}),c=ie.hasEagerState?ie.eagerState:n(c,ie.action);else{var ye={lane:_e,action:ie.action,hasEagerState:ie.hasEagerState,eagerState:ie.eagerState,next:null};B===null?(I=B=ye,M=c):B=B.next=ye,zt.lanes|=_e,kr|=_e}ie=ie.next}while(ie!==null&&ie!==g);B===null?M=c:B.next=I,ei(c,i.memoizedState)||(wn=!0),i.memoizedState=c,i.baseState=M,i.baseQueue=B,o.lastRenderedState=c}if(n=o.interleaved,n!==null){d=n;do g=d.lane,zt.lanes|=g,kr|=g,d=d.next;while(d!==n)}else d===null&&(o.lanes=0);return[i.memoizedState,o.dispatch]}function iu(n){var i=Yn(),o=i.queue;if(o===null)throw Error(t(311));o.lastRenderedReducer=n;var c=o.dispatch,d=o.pending,g=i.memoizedState;if(d!==null){o.pending=null;var M=d=d.next;do g=n(g,M.action),M=M.next;while(M!==d);ei(g,i.memoizedState)||(wn=!0),i.memoizedState=g,i.baseQueue===null&&(i.baseState=g),o.lastRenderedState=g}return[g,c]}function zh(){}function Hh(n,i){var o=zt,c=Yn(),d=i(),g=!ei(c.memoizedState,d);if(g&&(c.memoizedState=d,wn=!0),c=c.queue,ru(Wh.bind(null,o,c,n),[n]),c.getSnapshot!==i||g||nn!==null&&nn.memoizedState.tag&1){if(o.flags|=2048,Oo(9,Gh.bind(null,o,c,d,i),void 0,null),rn===null)throw Error(t(349));(Or&30)!==0||Vh(o,i,d)}return d}function Vh(n,i,o){n.flags|=16384,n={getSnapshot:i,value:o},i=zt.updateQueue,i===null?(i={lastEffect:null,stores:null},zt.updateQueue=i,i.stores=[n]):(o=i.stores,o===null?i.stores=[n]:o.push(n))}function Gh(n,i,o,c){i.value=o,i.getSnapshot=c,jh(i)&&Xh(n)}function Wh(n,i,o){return o(function(){jh(i)&&Xh(n)})}function jh(n){var i=n.getSnapshot;n=n.value;try{var o=i();return!ei(n,o)}catch{return!0}}function Xh(n){var i=bi(n,1);i!==null&&si(i,n,1,-1)}function Yh(n){var i=gi();return typeof n=="function"&&(n=n()),i.memoizedState=i.baseState=n,n={pending:null,interleaved:null,lanes:0,dispatch:null,lastRenderedReducer:Fo,lastRenderedState:n},i.queue=n,n=n.dispatch=_v.bind(null,zt,n),[i.memoizedState,n]}function Oo(n,i,o,c){return n={tag:n,create:i,destroy:o,deps:c,next:null},i=zt.updateQueue,i===null?(i={lastEffect:null,stores:null},zt.updateQueue=i,i.lastEffect=n.next=n):(o=i.lastEffect,o===null?i.lastEffect=n.next=n:(c=o.next,o.next=n,n.next=c,i.lastEffect=n)),n}function qh(){return Yn().memoizedState}function Wa(n,i,o,c){var d=gi();zt.flags|=n,d.memoizedState=Oo(1|i,o,void 0,c===void 0?null:c)}function ja(n,i,o,c){var d=Yn();c=c===void 0?null:c;var g=void 0;if($t!==null){var M=$t.memoizedState;if(g=M.destroy,c!==null&&Jc(c,M.deps)){d.memoizedState=Oo(i,o,g,c);return}}zt.flags|=n,d.memoizedState=Oo(1|i,o,g,c)}function $h(n,i){return Wa(8390656,8,n,i)}function ru(n,i){return ja(2048,8,n,i)}function Kh(n,i){return ja(4,2,n,i)}function Zh(n,i){return ja(4,4,n,i)}function Qh(n,i){if(typeof i=="function")return n=n(),i(n),function(){i(null)};if(i!=null)return n=n(),i.current=n,function(){i.current=null}}function Jh(n,i,o){return o=o!=null?o.concat([n]):null,ja(4,4,Qh.bind(null,i,n),o)}function su(){}function ep(n,i){var o=Yn();i=i===void 0?null:i;var c=o.memoizedState;return c!==null&&i!==null&&Jc(i,c[1])?c[0]:(o.memoizedState=[n,i],n)}function tp(n,i){var o=Yn();i=i===void 0?null:i;var c=o.memoizedState;return c!==null&&i!==null&&Jc(i,c[1])?c[0]:(n=n(),o.memoizedState=[n,i],n)}function np(n,i,o){return(Or&21)===0?(n.baseState&&(n.baseState=!1,wn=!0),n.memoizedState=o):(ei(o,i)||(o=us(),zt.lanes|=o,kr|=o,n.baseState=!0),i)}function mv(n,i){var o=wt;wt=o!==0&&4>o?o:4,n(!0);var c=Qc.transition;Qc.transition={};try{n(!1),i()}finally{wt=o,Qc.transition=c}}function ip(){return Yn().memoizedState}function gv(n,i,o){var c=ur(n);if(o={lane:c,action:o,hasEagerState:!1,eagerState:null,next:null},rp(n))sp(i,o);else if(o=Rh(n,i,o,c),o!==null){var d=yn();si(o,n,c,d),op(o,i,c)}}function _v(n,i,o){var c=ur(n),d={lane:c,action:o,hasEagerState:!1,eagerState:null,next:null};if(rp(n))sp(i,d);else{var g=n.alternate;if(n.lanes===0&&(g===null||g.lanes===0)&&(g=i.lastRenderedReducer,g!==null))try{var M=i.lastRenderedState,I=g(M,o);if(d.hasEagerState=!0,d.eagerState=I,ei(I,M)){var B=i.interleaved;B===null?(d.next=d,Wc(i)):(d.next=B.next,B.next=d),i.interleaved=d;return}}catch{}o=Rh(n,i,d,c),o!==null&&(d=yn(),si(o,n,c,d),op(o,i,c))}}function rp(n){var i=n.alternate;return n===zt||i!==null&&i===zt}function sp(n,i){Io=Ga=!0;var o=n.pending;o===null?i.next=i:(i.next=o.next,o.next=i),n.pending=i}function op(n,i,o){if((o&4194240)!==0){var c=i.lanes;c&=n.pendingLanes,o|=c,i.lanes=o,sc(n,o)}}var Xa={readContext:Xn,useCallback:hn,useContext:hn,useEffect:hn,useImperativeHandle:hn,useInsertionEffect:hn,useLayoutEffect:hn,useMemo:hn,useReducer:hn,useRef:hn,useState:hn,useDebugValue:hn,useDeferredValue:hn,useTransition:hn,useMutableSource:hn,useSyncExternalStore:hn,useId:hn,unstable_isNewReconciler:!1},vv={readContext:Xn,useCallback:function(n,i){return gi().memoizedState=[n,i===void 0?null:i],n},useContext:Xn,useEffect:$h,useImperativeHandle:function(n,i,o){return o=o!=null?o.concat([n]):null,Wa(4194308,4,Qh.bind(null,i,n),o)},useLayoutEffect:function(n,i){return Wa(4194308,4,n,i)},useInsertionEffect:function(n,i){return Wa(4,2,n,i)},useMemo:function(n,i){var o=gi();return i=i===void 0?null:i,n=n(),o.memoizedState=[n,i],n},useReducer:function(n,i,o){var c=gi();return i=o!==void 0?o(i):i,c.memoizedState=c.baseState=i,n={pending:null,interleaved:null,lanes:0,dispatch:null,lastRenderedReducer:n,lastRenderedState:i},c.queue=n,n=n.dispatch=gv.bind(null,zt,n),[c.memoizedState,n]},useRef:function(n){var i=gi();return n={current:n},i.memoizedState=n},useState:Yh,useDebugValue:su,useDeferredValue:function(n){return gi().memoizedState=n},useTransition:function(){var n=Yh(!1),i=n[0];return n=mv.bind(null,n[1]),gi().memoizedState=n,[i,n]},useMutableSource:function(){},useSyncExternalStore:function(n,i,o){var c=zt,d=gi();if(Ot){if(o===void 0)throw Error(t(407));o=o()}else{if(o=i(),rn===null)throw Error(t(349));(Or&30)!==0||Vh(c,i,o)}d.memoizedState=o;var g={value:o,getSnapshot:i};return d.queue=g,$h(Wh.bind(null,c,g,n),[n]),c.flags|=2048,Oo(9,Gh.bind(null,c,g,o,i),void 0,null),o},useId:function(){var n=gi(),i=rn.identifierPrefix;if(Ot){var o=Ri,c=Ci;o=(c&~(1<<32-ke(c)-1)).toString(32)+o,i=":"+i+"R"+o,o=Uo++,0<o&&(i+="H"+o.toString(32)),i+=":"}else o=pv++,i=":"+i+"r"+o.toString(32)+":";return n.memoizedState=i},unstable_isNewReconciler:!1},xv={readContext:Xn,useCallback:ep,useContext:Xn,useEffect:ru,useImperativeHandle:Jh,useInsertionEffect:Kh,useLayoutEffect:Zh,useMemo:tp,useReducer:nu,useRef:qh,useState:function(){return nu(Fo)},useDebugValue:su,useDeferredValue:function(n){var i=Yn();return np(i,$t.memoizedState,n)},useTransition:function(){var n=nu(Fo)[0],i=Yn().memoizedState;return[n,i]},useMutableSource:zh,useSyncExternalStore:Hh,useId:ip,unstable_isNewReconciler:!1},yv={readContext:Xn,useCallback:ep,useContext:Xn,useEffect:ru,useImperativeHandle:Jh,useInsertionEffect:Kh,useLayoutEffect:Zh,useMemo:tp,useReducer:iu,useRef:qh,useState:function(){return iu(Fo)},useDebugValue:su,useDeferredValue:function(n){var i=Yn();return $t===null?i.memoizedState=n:np(i,$t.memoizedState,n)},useTransition:function(){var n=iu(Fo)[0],i=Yn().memoizedState;return[n,i]},useMutableSource:zh,useSyncExternalStore:Hh,useId:ip,unstable_isNewReconciler:!1};function Rs(n,i){try{var o="",c=i;do o+=fe(c),c=c.return;while(c);var d=o}catch(g){d=`
Error generating stack: `+g.message+`
`+g.stack}return{value:n,source:i,stack:d,digest:null}}function ou(n,i,o){return{value:n,source:null,stack:o??null,digest:i??null}}function au(n,i){try{console.error(i.value)}catch(o){setTimeout(function(){throw o})}}var Sv=typeof WeakMap=="function"?WeakMap:Map;function ap(n,i,o){o=Pi(-1,o),o.tag=3,o.payload={element:null};var c=i.value;return o.callback=function(){Ja||(Ja=!0,Mu=c),au(n,i)},o}function lp(n,i,o){o=Pi(-1,o),o.tag=3;var c=n.type.getDerivedStateFromError;if(typeof c=="function"){var d=i.value;o.payload=function(){return c(d)},o.callback=function(){au(n,i)}}var g=n.stateNode;return g!==null&&typeof g.componentDidCatch=="function"&&(o.callback=function(){au(n,i),typeof c!="function"&&(lr===null?lr=new Set([this]):lr.add(this));var M=i.stack;this.componentDidCatch(i.value,{componentStack:M!==null?M:""})}),o}function cp(n,i,o){var c=n.pingCache;if(c===null){c=n.pingCache=new Sv;var d=new Set;c.set(i,d)}else d=c.get(i),d===void 0&&(d=new Set,c.set(i,d));d.has(o)||(d.add(o),n=Uv.bind(null,n,i,o),i.then(n,n))}function up(n){do{var i;if((i=n.tag===13)&&(i=n.memoizedState,i=i!==null?i.dehydrated!==null:!0),i)return n;n=n.return}while(n!==null);return null}function fp(n,i,o,c,d){return(n.mode&1)===0?(n===i?n.flags|=65536:(n.flags|=128,o.flags|=131072,o.flags&=-52805,o.tag===1&&(o.alternate===null?o.tag=17:(i=Pi(-1,1),i.tag=2,or(o,i,1))),o.lanes|=1),n):(n.flags|=65536,n.lanes=d,n)}var Mv=D.ReactCurrentOwner,wn=!1;function xn(n,i,o,c){i.child=n===null?kh(i,null,o,c):As(i,n.child,o,c)}function dp(n,i,o,c,d){o=o.render;var g=i.ref;return ws(i,d),c=eu(n,i,o,c,g,d),o=tu(),n!==null&&!wn?(i.updateQueue=n.updateQueue,i.flags&=-2053,n.lanes&=~d,Li(n,i,d)):(Ot&&o&&Uc(i),i.flags|=1,xn(n,i,c,d),i.child)}function hp(n,i,o,c,d){if(n===null){var g=o.type;return typeof g=="function"&&!bu(g)&&g.defaultProps===void 0&&o.compare===null&&o.defaultProps===void 0?(i.tag=15,i.type=g,pp(n,i,g,c,d)):(n=sl(o.type,null,c,i,i.mode,d),n.ref=i.ref,n.return=i,i.child=n)}if(g=n.child,(n.lanes&d)===0){var M=g.memoizedProps;if(o=o.compare,o=o!==null?o:Eo,o(M,c)&&n.ref===i.ref)return Li(n,i,d)}return i.flags|=1,n=dr(g,c),n.ref=i.ref,n.return=i,i.child=n}function pp(n,i,o,c,d){if(n!==null){var g=n.memoizedProps;if(Eo(g,c)&&n.ref===i.ref)if(wn=!1,i.pendingProps=c=g,(n.lanes&d)!==0)(n.flags&131072)!==0&&(wn=!0);else return i.lanes=n.lanes,Li(n,i,d)}return lu(n,i,o,c,d)}function mp(n,i,o){var c=i.pendingProps,d=c.children,g=n!==null?n.memoizedState:null;if(c.mode==="hidden")if((i.mode&1)===0)i.memoizedState={baseLanes:0,cachePool:null,transitions:null},Dt(Ps,kn),kn|=o;else{if((o&1073741824)===0)return n=g!==null?g.baseLanes|o:o,i.lanes=i.childLanes=1073741824,i.memoizedState={baseLanes:n,cachePool:null,transitions:null},i.updateQueue=null,Dt(Ps,kn),kn|=n,null;i.memoizedState={baseLanes:0,cachePool:null,transitions:null},c=g!==null?g.baseLanes:o,Dt(Ps,kn),kn|=c}else g!==null?(c=g.baseLanes|o,i.memoizedState=null):c=o,Dt(Ps,kn),kn|=c;return xn(n,i,d,o),i.child}function gp(n,i){var o=i.ref;(n===null&&o!==null||n!==null&&n.ref!==o)&&(i.flags|=512,i.flags|=2097152)}function lu(n,i,o,c,d){var g=Tn(o)?Dr:dn.current;return g=ys(i,g),ws(i,d),o=eu(n,i,o,c,g,d),c=tu(),n!==null&&!wn?(i.updateQueue=n.updateQueue,i.flags&=-2053,n.lanes&=~d,Li(n,i,d)):(Ot&&c&&Uc(i),i.flags|=1,xn(n,i,o,d),i.child)}function _p(n,i,o,c,d){if(Tn(o)){var g=!0;Pa(i)}else g=!1;if(ws(i,d),i.stateNode===null)qa(n,i),Ih(i,o,c),Yc(i,o,c,d),c=!0;else if(n===null){var M=i.stateNode,I=i.memoizedProps;M.props=I;var B=M.context,ie=o.contextType;typeof ie=="object"&&ie!==null?ie=Xn(ie):(ie=Tn(o)?Dr:dn.current,ie=ys(i,ie));var _e=o.getDerivedStateFromProps,ye=typeof _e=="function"||typeof M.getSnapshotBeforeUpdate=="function";ye||typeof M.UNSAFE_componentWillReceiveProps!="function"&&typeof M.componentWillReceiveProps!="function"||(I!==c||B!==ie)&&Uh(i,M,c,ie),sr=!1;var ge=i.memoizedState;M.state=ge,ka(i,c,M,d),B=i.memoizedState,I!==c||ge!==B||En.current||sr?(typeof _e=="function"&&(Xc(i,o,_e,c),B=i.memoizedState),(I=sr||Nh(i,o,I,c,ge,B,ie))?(ye||typeof M.UNSAFE_componentWillMount!="function"&&typeof M.componentWillMount!="function"||(typeof M.componentWillMount=="function"&&M.componentWillMount(),typeof M.UNSAFE_componentWillMount=="function"&&M.UNSAFE_componentWillMount()),typeof M.componentDidMount=="function"&&(i.flags|=4194308)):(typeof M.componentDidMount=="function"&&(i.flags|=4194308),i.memoizedProps=c,i.memoizedState=B),M.props=c,M.state=B,M.context=ie,c=I):(typeof M.componentDidMount=="function"&&(i.flags|=4194308),c=!1)}else{M=i.stateNode,bh(n,i),I=i.memoizedProps,ie=i.type===i.elementType?I:ni(i.type,I),M.props=ie,ye=i.pendingProps,ge=M.context,B=o.contextType,typeof B=="object"&&B!==null?B=Xn(B):(B=Tn(o)?Dr:dn.current,B=ys(i,B));var Ue=o.getDerivedStateFromProps;(_e=typeof Ue=="function"||typeof M.getSnapshotBeforeUpdate=="function")||typeof M.UNSAFE_componentWillReceiveProps!="function"&&typeof M.componentWillReceiveProps!="function"||(I!==ye||ge!==B)&&Uh(i,M,c,B),sr=!1,ge=i.memoizedState,M.state=ge,ka(i,c,M,d);var He=i.memoizedState;I!==ye||ge!==He||En.current||sr?(typeof Ue=="function"&&(Xc(i,o,Ue,c),He=i.memoizedState),(ie=sr||Nh(i,o,ie,c,ge,He,B)||!1)?(_e||typeof M.UNSAFE_componentWillUpdate!="function"&&typeof M.componentWillUpdate!="function"||(typeof M.componentWillUpdate=="function"&&M.componentWillUpdate(c,He,B),typeof M.UNSAFE_componentWillUpdate=="function"&&M.UNSAFE_componentWillUpdate(c,He,B)),typeof M.componentDidUpdate=="function"&&(i.flags|=4),typeof M.getSnapshotBeforeUpdate=="function"&&(i.flags|=1024)):(typeof M.componentDidUpdate!="function"||I===n.memoizedProps&&ge===n.memoizedState||(i.flags|=4),typeof M.getSnapshotBeforeUpdate!="function"||I===n.memoizedProps&&ge===n.memoizedState||(i.flags|=1024),i.memoizedProps=c,i.memoizedState=He),M.props=c,M.state=He,M.context=B,c=ie):(typeof M.componentDidUpdate!="function"||I===n.memoizedProps&&ge===n.memoizedState||(i.flags|=4),typeof M.getSnapshotBeforeUpdate!="function"||I===n.memoizedProps&&ge===n.memoizedState||(i.flags|=1024),c=!1)}return cu(n,i,o,c,g,d)}function cu(n,i,o,c,d,g){gp(n,i);var M=(i.flags&128)!==0;if(!c&&!M)return d&&Sh(i,o,!1),Li(n,i,g);c=i.stateNode,Mv.current=i;var I=M&&typeof o.getDerivedStateFromError!="function"?null:c.render();return i.flags|=1,n!==null&&M?(i.child=As(i,n.child,null,g),i.child=As(i,null,I,g)):xn(n,i,I,g),i.memoizedState=c.state,d&&Sh(i,o,!0),i.child}function vp(n){var i=n.stateNode;i.pendingContext?xh(n,i.pendingContext,i.pendingContext!==i.context):i.context&&xh(n,i.context,!1),qc(n,i.containerInfo)}function xp(n,i,o,c,d){return Es(),Bc(d),i.flags|=256,xn(n,i,o,c),i.child}var uu={dehydrated:null,treeContext:null,retryLane:0};function fu(n){return{baseLanes:n,cachePool:null,transitions:null}}function yp(n,i,o){var c=i.pendingProps,d=Bt.current,g=!1,M=(i.flags&128)!==0,I;if((I=M)||(I=n!==null&&n.memoizedState===null?!1:(d&2)!==0),I?(g=!0,i.flags&=-129):(n===null||n.memoizedState!==null)&&(d|=1),Dt(Bt,d&1),n===null)return kc(i),n=i.memoizedState,n!==null&&(n=n.dehydrated,n!==null)?((i.mode&1)===0?i.lanes=1:n.data==="$!"?i.lanes=8:i.lanes=1073741824,null):(M=c.children,n=c.fallback,g?(c=i.mode,g=i.child,M={mode:"hidden",children:M},(c&1)===0&&g!==null?(g.childLanes=0,g.pendingProps=M):g=ol(M,c,0,null),n=Vr(n,c,o,null),g.return=i,n.return=i,g.sibling=n,i.child=g,i.child.memoizedState=fu(o),i.memoizedState=uu,n):du(i,M));if(d=n.memoizedState,d!==null&&(I=d.dehydrated,I!==null))return Ev(n,i,M,c,I,d,o);if(g){g=c.fallback,M=i.mode,d=n.child,I=d.sibling;var B={mode:"hidden",children:c.children};return(M&1)===0&&i.child!==d?(c=i.child,c.childLanes=0,c.pendingProps=B,i.deletions=null):(c=dr(d,B),c.subtreeFlags=d.subtreeFlags&14680064),I!==null?g=dr(I,g):(g=Vr(g,M,o,null),g.flags|=2),g.return=i,c.return=i,c.sibling=g,i.child=c,c=g,g=i.child,M=n.child.memoizedState,M=M===null?fu(o):{baseLanes:M.baseLanes|o,cachePool:null,transitions:M.transitions},g.memoizedState=M,g.childLanes=n.childLanes&~o,i.memoizedState=uu,c}return g=n.child,n=g.sibling,c=dr(g,{mode:"visible",children:c.children}),(i.mode&1)===0&&(c.lanes=o),c.return=i,c.sibling=null,n!==null&&(o=i.deletions,o===null?(i.deletions=[n],i.flags|=16):o.push(n)),i.child=c,i.memoizedState=null,c}function du(n,i){return i=ol({mode:"visible",children:i},n.mode,0,null),i.return=n,n.child=i}function Ya(n,i,o,c){return c!==null&&Bc(c),As(i,n.child,null,o),n=du(i,i.pendingProps.children),n.flags|=2,i.memoizedState=null,n}function Ev(n,i,o,c,d,g,M){if(o)return i.flags&256?(i.flags&=-257,c=ou(Error(t(422))),Ya(n,i,M,c)):i.memoizedState!==null?(i.child=n.child,i.flags|=128,null):(g=c.fallback,d=i.mode,c=ol({mode:"visible",children:c.children},d,0,null),g=Vr(g,d,M,null),g.flags|=2,c.return=i,g.return=i,c.sibling=g,i.child=c,(i.mode&1)!==0&&As(i,n.child,null,M),i.child.memoizedState=fu(M),i.memoizedState=uu,g);if((i.mode&1)===0)return Ya(n,i,M,null);if(d.data==="$!"){if(c=d.nextSibling&&d.nextSibling.dataset,c)var I=c.dgst;return c=I,g=Error(t(419)),c=ou(g,c,void 0),Ya(n,i,M,c)}if(I=(M&n.childLanes)!==0,wn||I){if(c=rn,c!==null){switch(M&-M){case 4:d=2;break;case 16:d=8;break;case 64:case 128:case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:case 262144:case 524288:case 1048576:case 2097152:case 4194304:case 8388608:case 16777216:case 33554432:case 67108864:d=32;break;case 536870912:d=268435456;break;default:d=0}d=(d&(c.suspendedLanes|M))!==0?0:d,d!==0&&d!==g.retryLane&&(g.retryLane=d,bi(n,d),si(c,n,d,-1))}return Ru(),c=ou(Error(t(421))),Ya(n,i,M,c)}return d.data==="$?"?(i.flags|=128,i.child=n.child,i=Fv.bind(null,n),d._reactRetry=i,null):(n=g.treeContext,On=tr(d.nextSibling),Fn=i,Ot=!0,ti=null,n!==null&&(Wn[jn++]=Ci,Wn[jn++]=Ri,Wn[jn++]=Nr,Ci=n.id,Ri=n.overflow,Nr=i),i=du(i,c.children),i.flags|=4096,i)}function Sp(n,i,o){n.lanes|=i;var c=n.alternate;c!==null&&(c.lanes|=i),Gc(n.return,i,o)}function hu(n,i,o,c,d){var g=n.memoizedState;g===null?n.memoizedState={isBackwards:i,rendering:null,renderingStartTime:0,last:c,tail:o,tailMode:d}:(g.isBackwards=i,g.rendering=null,g.renderingStartTime=0,g.last=c,g.tail=o,g.tailMode=d)}function Mp(n,i,o){var c=i.pendingProps,d=c.revealOrder,g=c.tail;if(xn(n,i,c.children,o),c=Bt.current,(c&2)!==0)c=c&1|2,i.flags|=128;else{if(n!==null&&(n.flags&128)!==0)e:for(n=i.child;n!==null;){if(n.tag===13)n.memoizedState!==null&&Sp(n,o,i);else if(n.tag===19)Sp(n,o,i);else if(n.child!==null){n.child.return=n,n=n.child;continue}if(n===i)break e;for(;n.sibling===null;){if(n.return===null||n.return===i)break e;n=n.return}n.sibling.return=n.return,n=n.sibling}c&=1}if(Dt(Bt,c),(i.mode&1)===0)i.memoizedState=null;else switch(d){case"forwards":for(o=i.child,d=null;o!==null;)n=o.alternate,n!==null&&Ha(n)===null&&(d=o),o=o.sibling;o=d,o===null?(d=i.child,i.child=null):(d=o.sibling,o.sibling=null),hu(i,!1,d,o,g);break;case"backwards":for(o=null,d=i.child,i.child=null;d!==null;){if(n=d.alternate,n!==null&&Ha(n)===null){i.child=d;break}n=d.sibling,d.sibling=o,o=d,d=n}hu(i,!0,o,null,g);break;case"together":hu(i,!1,null,null,void 0);break;default:i.memoizedState=null}return i.child}function qa(n,i){(i.mode&1)===0&&n!==null&&(n.alternate=null,i.alternate=null,i.flags|=2)}function Li(n,i,o){if(n!==null&&(i.dependencies=n.dependencies),kr|=i.lanes,(o&i.childLanes)===0)return null;if(n!==null&&i.child!==n.child)throw Error(t(153));if(i.child!==null){for(n=i.child,o=dr(n,n.pendingProps),i.child=o,o.return=i;n.sibling!==null;)n=n.sibling,o=o.sibling=dr(n,n.pendingProps),o.return=i;o.sibling=null}return i.child}function Tv(n,i,o){switch(i.tag){case 3:vp(i),Es();break;case 5:Bh(i);break;case 1:Tn(i.type)&&Pa(i);break;case 4:qc(i,i.stateNode.containerInfo);break;case 10:var c=i.type._context,d=i.memoizedProps.value;Dt(Ua,c._currentValue),c._currentValue=d;break;case 13:if(c=i.memoizedState,c!==null)return c.dehydrated!==null?(Dt(Bt,Bt.current&1),i.flags|=128,null):(o&i.child.childLanes)!==0?yp(n,i,o):(Dt(Bt,Bt.current&1),n=Li(n,i,o),n!==null?n.sibling:null);Dt(Bt,Bt.current&1);break;case 19:if(c=(o&i.childLanes)!==0,(n.flags&128)!==0){if(c)return Mp(n,i,o);i.flags|=128}if(d=i.memoizedState,d!==null&&(d.rendering=null,d.tail=null,d.lastEffect=null),Dt(Bt,Bt.current),c)break;return null;case 22:case 23:return i.lanes=0,mp(n,i,o)}return Li(n,i,o)}var Ep,pu,Tp,wp;Ep=function(n,i){for(var o=i.child;o!==null;){if(o.tag===5||o.tag===6)n.appendChild(o.stateNode);else if(o.tag!==4&&o.child!==null){o.child.return=o,o=o.child;continue}if(o===i)break;for(;o.sibling===null;){if(o.return===null||o.return===i)return;o=o.return}o.sibling.return=o.return,o=o.sibling}},pu=function(){},Tp=function(n,i,o,c){var d=n.memoizedProps;if(d!==c){n=i.stateNode,Fr(mi.current);var g=null;switch(o){case"input":d=an(n,d),c=an(n,c),g=[];break;case"select":d=L({},d,{value:void 0}),c=L({},c,{value:void 0}),g=[];break;case"textarea":d=K(n,d),c=K(n,c),g=[];break;default:typeof d.onClick!="function"&&typeof c.onClick=="function"&&(n.onclick=Ca)}it(o,c);var M;o=null;for(ie in d)if(!c.hasOwnProperty(ie)&&d.hasOwnProperty(ie)&&d[ie]!=null)if(ie==="style"){var I=d[ie];for(M in I)I.hasOwnProperty(M)&&(o||(o={}),o[M]="")}else ie!=="dangerouslySetInnerHTML"&&ie!=="children"&&ie!=="suppressContentEditableWarning"&&ie!=="suppressHydrationWarning"&&ie!=="autoFocus"&&(a.hasOwnProperty(ie)?g||(g=[]):(g=g||[]).push(ie,null));for(ie in c){var B=c[ie];if(I=d?.[ie],c.hasOwnProperty(ie)&&B!==I&&(B!=null||I!=null))if(ie==="style")if(I){for(M in I)!I.hasOwnProperty(M)||B&&B.hasOwnProperty(M)||(o||(o={}),o[M]="");for(M in B)B.hasOwnProperty(M)&&I[M]!==B[M]&&(o||(o={}),o[M]=B[M])}else o||(g||(g=[]),g.push(ie,o)),o=B;else ie==="dangerouslySetInnerHTML"?(B=B?B.__html:void 0,I=I?I.__html:void 0,B!=null&&I!==B&&(g=g||[]).push(ie,B)):ie==="children"?typeof B!="string"&&typeof B!="number"||(g=g||[]).push(ie,""+B):ie!=="suppressContentEditableWarning"&&ie!=="suppressHydrationWarning"&&(a.hasOwnProperty(ie)?(B!=null&&ie==="onScroll"&&Nt("scroll",n),g||I===B||(g=[])):(g=g||[]).push(ie,B))}o&&(g=g||[]).push("style",o);var ie=g;(i.updateQueue=ie)&&(i.flags|=4)}},wp=function(n,i,o,c){o!==c&&(i.flags|=4)};function ko(n,i){if(!Ot)switch(n.tailMode){case"hidden":i=n.tail;for(var o=null;i!==null;)i.alternate!==null&&(o=i),i=i.sibling;o===null?n.tail=null:o.sibling=null;break;case"collapsed":o=n.tail;for(var c=null;o!==null;)o.alternate!==null&&(c=o),o=o.sibling;c===null?i||n.tail===null?n.tail=null:n.tail.sibling=null:c.sibling=null}}function pn(n){var i=n.alternate!==null&&n.alternate.child===n.child,o=0,c=0;if(i)for(var d=n.child;d!==null;)o|=d.lanes|d.childLanes,c|=d.subtreeFlags&14680064,c|=d.flags&14680064,d.return=n,d=d.sibling;else for(d=n.child;d!==null;)o|=d.lanes|d.childLanes,c|=d.subtreeFlags,c|=d.flags,d.return=n,d=d.sibling;return n.subtreeFlags|=c,n.childLanes=o,i}function wv(n,i,o){var c=i.pendingProps;switch(Fc(i),i.tag){case 2:case 16:case 15:case 0:case 11:case 7:case 8:case 12:case 9:case 14:return pn(i),null;case 1:return Tn(i.type)&&ba(),pn(i),null;case 3:return c=i.stateNode,Cs(),It(En),It(dn),Zc(),c.pendingContext&&(c.context=c.pendingContext,c.pendingContext=null),(n===null||n.child===null)&&(Ia(i)?i.flags|=4:n===null||n.memoizedState.isDehydrated&&(i.flags&256)===0||(i.flags|=1024,ti!==null&&(wu(ti),ti=null))),pu(n,i),pn(i),null;case 5:$c(i);var d=Fr(No.current);if(o=i.type,n!==null&&i.stateNode!=null)Tp(n,i,o,c,d),n.ref!==i.ref&&(i.flags|=512,i.flags|=2097152);else{if(!c){if(i.stateNode===null)throw Error(t(166));return pn(i),null}if(n=Fr(mi.current),Ia(i)){c=i.stateNode,o=i.type;var g=i.memoizedProps;switch(c[pi]=i,c[Ro]=g,n=(i.mode&1)!==0,o){case"dialog":Nt("cancel",c),Nt("close",c);break;case"iframe":case"object":case"embed":Nt("load",c);break;case"video":case"audio":for(d=0;d<wo.length;d++)Nt(wo[d],c);break;case"source":Nt("error",c);break;case"img":case"image":case"link":Nt("error",c),Nt("load",c);break;case"details":Nt("toggle",c);break;case"input":at(c,g),Nt("invalid",c);break;case"select":c._wrapperState={wasMultiple:!!g.multiple},Nt("invalid",c);break;case"textarea":pe(c,g),Nt("invalid",c)}it(o,g),d=null;for(var M in g)if(g.hasOwnProperty(M)){var I=g[M];M==="children"?typeof I=="string"?c.textContent!==I&&(g.suppressHydrationWarning!==!0&&Aa(c.textContent,I,n),d=["children",I]):typeof I=="number"&&c.textContent!==""+I&&(g.suppressHydrationWarning!==!0&&Aa(c.textContent,I,n),d=["children",""+I]):a.hasOwnProperty(M)&&I!=null&&M==="onScroll"&&Nt("scroll",c)}switch(o){case"input":gt(c),At(c,g,!0);break;case"textarea":gt(c),de(c);break;case"select":case"option":break;default:typeof g.onClick=="function"&&(c.onclick=Ca)}c=d,i.updateQueue=c,c!==null&&(i.flags|=4)}else{M=d.nodeType===9?d:d.ownerDocument,n==="http://www.w3.org/1999/xhtml"&&(n=Ye(o)),n==="http://www.w3.org/1999/xhtml"?o==="script"?(n=M.createElement("div"),n.innerHTML="<script><\/script>",n=n.removeChild(n.firstChild)):typeof c.is=="string"?n=M.createElement(o,{is:c.is}):(n=M.createElement(o),o==="select"&&(M=n,c.multiple?M.multiple=!0:c.size&&(M.size=c.size))):n=M.createElementNS(n,o),n[pi]=i,n[Ro]=c,Ep(n,i,!1,!1),i.stateNode=n;e:{switch(M=Et(o,c),o){case"dialog":Nt("cancel",n),Nt("close",n),d=c;break;case"iframe":case"object":case"embed":Nt("load",n),d=c;break;case"video":case"audio":for(d=0;d<wo.length;d++)Nt(wo[d],n);d=c;break;case"source":Nt("error",n),d=c;break;case"img":case"image":case"link":Nt("error",n),Nt("load",n),d=c;break;case"details":Nt("toggle",n),d=c;break;case"input":at(n,c),d=an(n,c),Nt("invalid",n);break;case"option":d=c;break;case"select":n._wrapperState={wasMultiple:!!c.multiple},d=L({},c,{value:void 0}),Nt("invalid",n);break;case"textarea":pe(n,c),d=K(n,c),Nt("invalid",n);break;default:d=c}it(o,d),I=d;for(g in I)if(I.hasOwnProperty(g)){var B=I[g];g==="style"?Be(n,B):g==="dangerouslySetInnerHTML"?(B=B?B.__html:void 0,B!=null&&pt(n,B)):g==="children"?typeof B=="string"?(o!=="textarea"||B!=="")&&Ee(n,B):typeof B=="number"&&Ee(n,""+B):g!=="suppressContentEditableWarning"&&g!=="suppressHydrationWarning"&&g!=="autoFocus"&&(a.hasOwnProperty(g)?B!=null&&g==="onScroll"&&Nt("scroll",n):B!=null&&b(n,g,B,M))}switch(o){case"input":gt(n),At(n,c,!1);break;case"textarea":gt(n),de(n);break;case"option":c.value!=null&&n.setAttribute("value",""+Pe(c.value));break;case"select":n.multiple=!!c.multiple,g=c.value,g!=null?A(n,!!c.multiple,g,!1):c.defaultValue!=null&&A(n,!!c.multiple,c.defaultValue,!0);break;default:typeof d.onClick=="function"&&(n.onclick=Ca)}switch(o){case"button":case"input":case"select":case"textarea":c=!!c.autoFocus;break e;case"img":c=!0;break e;default:c=!1}}c&&(i.flags|=4)}i.ref!==null&&(i.flags|=512,i.flags|=2097152)}return pn(i),null;case 6:if(n&&i.stateNode!=null)wp(n,i,n.memoizedProps,c);else{if(typeof c!="string"&&i.stateNode===null)throw Error(t(166));if(o=Fr(No.current),Fr(mi.current),Ia(i)){if(c=i.stateNode,o=i.memoizedProps,c[pi]=i,(g=c.nodeValue!==o)&&(n=Fn,n!==null))switch(n.tag){case 3:Aa(c.nodeValue,o,(n.mode&1)!==0);break;case 5:n.memoizedProps.suppressHydrationWarning!==!0&&Aa(c.nodeValue,o,(n.mode&1)!==0)}g&&(i.flags|=4)}else c=(o.nodeType===9?o:o.ownerDocument).createTextNode(c),c[pi]=i,i.stateNode=c}return pn(i),null;case 13:if(It(Bt),c=i.memoizedState,n===null||n.memoizedState!==null&&n.memoizedState.dehydrated!==null){if(Ot&&On!==null&&(i.mode&1)!==0&&(i.flags&128)===0)Ch(),Es(),i.flags|=98560,g=!1;else if(g=Ia(i),c!==null&&c.dehydrated!==null){if(n===null){if(!g)throw Error(t(318));if(g=i.memoizedState,g=g!==null?g.dehydrated:null,!g)throw Error(t(317));g[pi]=i}else Es(),(i.flags&128)===0&&(i.memoizedState=null),i.flags|=4;pn(i),g=!1}else ti!==null&&(wu(ti),ti=null),g=!0;if(!g)return i.flags&65536?i:null}return(i.flags&128)!==0?(i.lanes=o,i):(c=c!==null,c!==(n!==null&&n.memoizedState!==null)&&c&&(i.child.flags|=8192,(i.mode&1)!==0&&(n===null||(Bt.current&1)!==0?Kt===0&&(Kt=3):Ru())),i.updateQueue!==null&&(i.flags|=4),pn(i),null);case 4:return Cs(),pu(n,i),n===null&&Ao(i.stateNode.containerInfo),pn(i),null;case 10:return Vc(i.type._context),pn(i),null;case 17:return Tn(i.type)&&ba(),pn(i),null;case 19:if(It(Bt),g=i.memoizedState,g===null)return pn(i),null;if(c=(i.flags&128)!==0,M=g.rendering,M===null)if(c)ko(g,!1);else{if(Kt!==0||n!==null&&(n.flags&128)!==0)for(n=i.child;n!==null;){if(M=Ha(n),M!==null){for(i.flags|=128,ko(g,!1),c=M.updateQueue,c!==null&&(i.updateQueue=c,i.flags|=4),i.subtreeFlags=0,c=o,o=i.child;o!==null;)g=o,n=c,g.flags&=14680066,M=g.alternate,M===null?(g.childLanes=0,g.lanes=n,g.child=null,g.subtreeFlags=0,g.memoizedProps=null,g.memoizedState=null,g.updateQueue=null,g.dependencies=null,g.stateNode=null):(g.childLanes=M.childLanes,g.lanes=M.lanes,g.child=M.child,g.subtreeFlags=0,g.deletions=null,g.memoizedProps=M.memoizedProps,g.memoizedState=M.memoizedState,g.updateQueue=M.updateQueue,g.type=M.type,n=M.dependencies,g.dependencies=n===null?null:{lanes:n.lanes,firstContext:n.firstContext}),o=o.sibling;return Dt(Bt,Bt.current&1|2),i.child}n=n.sibling}g.tail!==null&&Ae()>Ls&&(i.flags|=128,c=!0,ko(g,!1),i.lanes=4194304)}else{if(!c)if(n=Ha(M),n!==null){if(i.flags|=128,c=!0,o=n.updateQueue,o!==null&&(i.updateQueue=o,i.flags|=4),ko(g,!0),g.tail===null&&g.tailMode==="hidden"&&!M.alternate&&!Ot)return pn(i),null}else 2*Ae()-g.renderingStartTime>Ls&&o!==1073741824&&(i.flags|=128,c=!0,ko(g,!1),i.lanes=4194304);g.isBackwards?(M.sibling=i.child,i.child=M):(o=g.last,o!==null?o.sibling=M:i.child=M,g.last=M)}return g.tail!==null?(i=g.tail,g.rendering=i,g.tail=i.sibling,g.renderingStartTime=Ae(),i.sibling=null,o=Bt.current,Dt(Bt,c?o&1|2:o&1),i):(pn(i),null);case 22:case 23:return Cu(),c=i.memoizedState!==null,n!==null&&n.memoizedState!==null!==c&&(i.flags|=8192),c&&(i.mode&1)!==0?(kn&1073741824)!==0&&(pn(i),i.subtreeFlags&6&&(i.flags|=8192)):pn(i),null;case 24:return null;case 25:return null}throw Error(t(156,i.tag))}function Av(n,i){switch(Fc(i),i.tag){case 1:return Tn(i.type)&&ba(),n=i.flags,n&65536?(i.flags=n&-65537|128,i):null;case 3:return Cs(),It(En),It(dn),Zc(),n=i.flags,(n&65536)!==0&&(n&128)===0?(i.flags=n&-65537|128,i):null;case 5:return $c(i),null;case 13:if(It(Bt),n=i.memoizedState,n!==null&&n.dehydrated!==null){if(i.alternate===null)throw Error(t(340));Es()}return n=i.flags,n&65536?(i.flags=n&-65537|128,i):null;case 19:return It(Bt),null;case 4:return Cs(),null;case 10:return Vc(i.type._context),null;case 22:case 23:return Cu(),null;case 24:return null;default:return null}}var $a=!1,mn=!1,Cv=typeof WeakSet=="function"?WeakSet:Set,ze=null;function bs(n,i){var o=n.ref;if(o!==null)if(typeof o=="function")try{o(null)}catch(c){Ht(n,i,c)}else o.current=null}function mu(n,i,o){try{o()}catch(c){Ht(n,i,c)}}var Ap=!1;function Rv(n,i){if(Cc=ma,n=ih(),xc(n)){if("selectionStart"in n)var o={start:n.selectionStart,end:n.selectionEnd};else e:{o=(o=n.ownerDocument)&&o.defaultView||window;var c=o.getSelection&&o.getSelection();if(c&&c.rangeCount!==0){o=c.anchorNode;var d=c.anchorOffset,g=c.focusNode;c=c.focusOffset;try{o.nodeType,g.nodeType}catch{o=null;break e}var M=0,I=-1,B=-1,ie=0,_e=0,ye=n,ge=null;t:for(;;){for(var Ue;ye!==o||d!==0&&ye.nodeType!==3||(I=M+d),ye!==g||c!==0&&ye.nodeType!==3||(B=M+c),ye.nodeType===3&&(M+=ye.nodeValue.length),(Ue=ye.firstChild)!==null;)ge=ye,ye=Ue;for(;;){if(ye===n)break t;if(ge===o&&++ie===d&&(I=M),ge===g&&++_e===c&&(B=M),(Ue=ye.nextSibling)!==null)break;ye=ge,ge=ye.parentNode}ye=Ue}o=I===-1||B===-1?null:{start:I,end:B}}else o=null}o=o||{start:0,end:0}}else o=null;for(Rc={focusedElem:n,selectionRange:o},ma=!1,ze=i;ze!==null;)if(i=ze,n=i.child,(i.subtreeFlags&1028)!==0&&n!==null)n.return=i,ze=n;else for(;ze!==null;){i=ze;try{var He=i.alternate;if((i.flags&1024)!==0)switch(i.tag){case 0:case 11:case 15:break;case 1:if(He!==null){var Ve=He.memoizedProps,Wt=He.memoizedState,q=i.stateNode,H=q.getSnapshotBeforeUpdate(i.elementType===i.type?Ve:ni(i.type,Ve),Wt);q.__reactInternalSnapshotBeforeUpdate=H}break;case 3:var $=i.stateNode.containerInfo;$.nodeType===1?$.textContent="":$.nodeType===9&&$.documentElement&&$.removeChild($.documentElement);break;case 5:case 6:case 4:case 17:break;default:throw Error(t(163))}}catch(Te){Ht(i,i.return,Te)}if(n=i.sibling,n!==null){n.return=i.return,ze=n;break}ze=i.return}return He=Ap,Ap=!1,He}function Bo(n,i,o){var c=i.updateQueue;if(c=c!==null?c.lastEffect:null,c!==null){var d=c=c.next;do{if((d.tag&n)===n){var g=d.destroy;d.destroy=void 0,g!==void 0&&mu(i,o,g)}d=d.next}while(d!==c)}}function Ka(n,i){if(i=i.updateQueue,i=i!==null?i.lastEffect:null,i!==null){var o=i=i.next;do{if((o.tag&n)===n){var c=o.create;o.destroy=c()}o=o.next}while(o!==i)}}function gu(n){var i=n.ref;if(i!==null){var o=n.stateNode;n.tag,n=o,typeof i=="function"?i(n):i.current=n}}function Cp(n){var i=n.alternate;i!==null&&(n.alternate=null,Cp(i)),n.child=null,n.deletions=null,n.sibling=null,n.tag===5&&(i=n.stateNode,i!==null&&(delete i[pi],delete i[Ro],delete i[Dc],delete i[uv],delete i[fv])),n.stateNode=null,n.return=null,n.dependencies=null,n.memoizedProps=null,n.memoizedState=null,n.pendingProps=null,n.stateNode=null,n.updateQueue=null}function Rp(n){return n.tag===5||n.tag===3||n.tag===4}function bp(n){e:for(;;){for(;n.sibling===null;){if(n.return===null||Rp(n.return))return null;n=n.return}for(n.sibling.return=n.return,n=n.sibling;n.tag!==5&&n.tag!==6&&n.tag!==18;){if(n.flags&2||n.child===null||n.tag===4)continue e;n.child.return=n,n=n.child}if(!(n.flags&2))return n.stateNode}}function _u(n,i,o){var c=n.tag;if(c===5||c===6)n=n.stateNode,i?o.nodeType===8?o.parentNode.insertBefore(n,i):o.insertBefore(n,i):(o.nodeType===8?(i=o.parentNode,i.insertBefore(n,o)):(i=o,i.appendChild(n)),o=o._reactRootContainer,o!=null||i.onclick!==null||(i.onclick=Ca));else if(c!==4&&(n=n.child,n!==null))for(_u(n,i,o),n=n.sibling;n!==null;)_u(n,i,o),n=n.sibling}function vu(n,i,o){var c=n.tag;if(c===5||c===6)n=n.stateNode,i?o.insertBefore(n,i):o.appendChild(n);else if(c!==4&&(n=n.child,n!==null))for(vu(n,i,o),n=n.sibling;n!==null;)vu(n,i,o),n=n.sibling}var cn=null,ii=!1;function ar(n,i,o){for(o=o.child;o!==null;)Pp(n,i,o),o=o.sibling}function Pp(n,i,o){if(Ft&&typeof Ft.onCommitFiberUnmount=="function")try{Ft.onCommitFiberUnmount(bt,o)}catch{}switch(o.tag){case 5:mn||bs(o,i);case 6:var c=cn,d=ii;cn=null,ar(n,i,o),cn=c,ii=d,cn!==null&&(ii?(n=cn,o=o.stateNode,n.nodeType===8?n.parentNode.removeChild(o):n.removeChild(o)):cn.removeChild(o.stateNode));break;case 18:cn!==null&&(ii?(n=cn,o=o.stateNode,n.nodeType===8?Lc(n.parentNode,o):n.nodeType===1&&Lc(n,o),_o(n)):Lc(cn,o.stateNode));break;case 4:c=cn,d=ii,cn=o.stateNode.containerInfo,ii=!0,ar(n,i,o),cn=c,ii=d;break;case 0:case 11:case 14:case 15:if(!mn&&(c=o.updateQueue,c!==null&&(c=c.lastEffect,c!==null))){d=c=c.next;do{var g=d,M=g.destroy;g=g.tag,M!==void 0&&((g&2)!==0||(g&4)!==0)&&mu(o,i,M),d=d.next}while(d!==c)}ar(n,i,o);break;case 1:if(!mn&&(bs(o,i),c=o.stateNode,typeof c.componentWillUnmount=="function"))try{c.props=o.memoizedProps,c.state=o.memoizedState,c.componentWillUnmount()}catch(I){Ht(o,i,I)}ar(n,i,o);break;case 21:ar(n,i,o);break;case 22:o.mode&1?(mn=(c=mn)||o.memoizedState!==null,ar(n,i,o),mn=c):ar(n,i,o);break;default:ar(n,i,o)}}function Lp(n){var i=n.updateQueue;if(i!==null){n.updateQueue=null;var o=n.stateNode;o===null&&(o=n.stateNode=new Cv),i.forEach(function(c){var d=Ov.bind(null,n,c);o.has(c)||(o.add(c),c.then(d,d))})}}function ri(n,i){var o=i.deletions;if(o!==null)for(var c=0;c<o.length;c++){var d=o[c];try{var g=n,M=i,I=M;e:for(;I!==null;){switch(I.tag){case 5:cn=I.stateNode,ii=!1;break e;case 3:cn=I.stateNode.containerInfo,ii=!0;break e;case 4:cn=I.stateNode.containerInfo,ii=!0;break e}I=I.return}if(cn===null)throw Error(t(160));Pp(g,M,d),cn=null,ii=!1;var B=d.alternate;B!==null&&(B.return=null),d.return=null}catch(ie){Ht(d,i,ie)}}if(i.subtreeFlags&12854)for(i=i.child;i!==null;)Dp(i,n),i=i.sibling}function Dp(n,i){var o=n.alternate,c=n.flags;switch(n.tag){case 0:case 11:case 14:case 15:if(ri(i,n),_i(n),c&4){try{Bo(3,n,n.return),Ka(3,n)}catch(Ve){Ht(n,n.return,Ve)}try{Bo(5,n,n.return)}catch(Ve){Ht(n,n.return,Ve)}}break;case 1:ri(i,n),_i(n),c&512&&o!==null&&bs(o,o.return);break;case 5:if(ri(i,n),_i(n),c&512&&o!==null&&bs(o,o.return),n.flags&32){var d=n.stateNode;try{Ee(d,"")}catch(Ve){Ht(n,n.return,Ve)}}if(c&4&&(d=n.stateNode,d!=null)){var g=n.memoizedProps,M=o!==null?o.memoizedProps:g,I=n.type,B=n.updateQueue;if(n.updateQueue=null,B!==null)try{I==="input"&&g.type==="radio"&&g.name!=null&&ht(d,g),Et(I,M);var ie=Et(I,g);for(M=0;M<B.length;M+=2){var _e=B[M],ye=B[M+1];_e==="style"?Be(d,ye):_e==="dangerouslySetInnerHTML"?pt(d,ye):_e==="children"?Ee(d,ye):b(d,_e,ye,ie)}switch(I){case"input":Ze(d,g);break;case"textarea":xe(d,g);break;case"select":var ge=d._wrapperState.wasMultiple;d._wrapperState.wasMultiple=!!g.multiple;var Ue=g.value;Ue!=null?A(d,!!g.multiple,Ue,!1):ge!==!!g.multiple&&(g.defaultValue!=null?A(d,!!g.multiple,g.defaultValue,!0):A(d,!!g.multiple,g.multiple?[]:"",!1))}d[Ro]=g}catch(Ve){Ht(n,n.return,Ve)}}break;case 6:if(ri(i,n),_i(n),c&4){if(n.stateNode===null)throw Error(t(162));d=n.stateNode,g=n.memoizedProps;try{d.nodeValue=g}catch(Ve){Ht(n,n.return,Ve)}}break;case 3:if(ri(i,n),_i(n),c&4&&o!==null&&o.memoizedState.isDehydrated)try{_o(i.containerInfo)}catch(Ve){Ht(n,n.return,Ve)}break;case 4:ri(i,n),_i(n);break;case 13:ri(i,n),_i(n),d=n.child,d.flags&8192&&(g=d.memoizedState!==null,d.stateNode.isHidden=g,!g||d.alternate!==null&&d.alternate.memoizedState!==null||(Su=Ae())),c&4&&Lp(n);break;case 22:if(_e=o!==null&&o.memoizedState!==null,n.mode&1?(mn=(ie=mn)||_e,ri(i,n),mn=ie):ri(i,n),_i(n),c&8192){if(ie=n.memoizedState!==null,(n.stateNode.isHidden=ie)&&!_e&&(n.mode&1)!==0)for(ze=n,_e=n.child;_e!==null;){for(ye=ze=_e;ze!==null;){switch(ge=ze,Ue=ge.child,ge.tag){case 0:case 11:case 14:case 15:Bo(4,ge,ge.return);break;case 1:bs(ge,ge.return);var He=ge.stateNode;if(typeof He.componentWillUnmount=="function"){c=ge,o=ge.return;try{i=c,He.props=i.memoizedProps,He.state=i.memoizedState,He.componentWillUnmount()}catch(Ve){Ht(c,o,Ve)}}break;case 5:bs(ge,ge.return);break;case 22:if(ge.memoizedState!==null){Up(ye);continue}}Ue!==null?(Ue.return=ge,ze=Ue):Up(ye)}_e=_e.sibling}e:for(_e=null,ye=n;;){if(ye.tag===5){if(_e===null){_e=ye;try{d=ye.stateNode,ie?(g=d.style,typeof g.setProperty=="function"?g.setProperty("display","none","important"):g.display="none"):(I=ye.stateNode,B=ye.memoizedProps.style,M=B!=null&&B.hasOwnProperty("display")?B.display:null,I.style.display=et("display",M))}catch(Ve){Ht(n,n.return,Ve)}}}else if(ye.tag===6){if(_e===null)try{ye.stateNode.nodeValue=ie?"":ye.memoizedProps}catch(Ve){Ht(n,n.return,Ve)}}else if((ye.tag!==22&&ye.tag!==23||ye.memoizedState===null||ye===n)&&ye.child!==null){ye.child.return=ye,ye=ye.child;continue}if(ye===n)break e;for(;ye.sibling===null;){if(ye.return===null||ye.return===n)break e;_e===ye&&(_e=null),ye=ye.return}_e===ye&&(_e=null),ye.sibling.return=ye.return,ye=ye.sibling}}break;case 19:ri(i,n),_i(n),c&4&&Lp(n);break;case 21:break;default:ri(i,n),_i(n)}}function _i(n){var i=n.flags;if(i&2){try{e:{for(var o=n.return;o!==null;){if(Rp(o)){var c=o;break e}o=o.return}throw Error(t(160))}switch(c.tag){case 5:var d=c.stateNode;c.flags&32&&(Ee(d,""),c.flags&=-33);var g=bp(n);vu(n,g,d);break;case 3:case 4:var M=c.stateNode.containerInfo,I=bp(n);_u(n,I,M);break;default:throw Error(t(161))}}catch(B){Ht(n,n.return,B)}n.flags&=-3}i&4096&&(n.flags&=-4097)}function bv(n,i,o){ze=n,Np(n)}function Np(n,i,o){for(var c=(n.mode&1)!==0;ze!==null;){var d=ze,g=d.child;if(d.tag===22&&c){var M=d.memoizedState!==null||$a;if(!M){var I=d.alternate,B=I!==null&&I.memoizedState!==null||mn;I=$a;var ie=mn;if($a=M,(mn=B)&&!ie)for(ze=d;ze!==null;)M=ze,B=M.child,M.tag===22&&M.memoizedState!==null?Fp(d):B!==null?(B.return=M,ze=B):Fp(d);for(;g!==null;)ze=g,Np(g),g=g.sibling;ze=d,$a=I,mn=ie}Ip(n)}else(d.subtreeFlags&8772)!==0&&g!==null?(g.return=d,ze=g):Ip(n)}}function Ip(n){for(;ze!==null;){var i=ze;if((i.flags&8772)!==0){var o=i.alternate;try{if((i.flags&8772)!==0)switch(i.tag){case 0:case 11:case 15:mn||Ka(5,i);break;case 1:var c=i.stateNode;if(i.flags&4&&!mn)if(o===null)c.componentDidMount();else{var d=i.elementType===i.type?o.memoizedProps:ni(i.type,o.memoizedProps);c.componentDidUpdate(d,o.memoizedState,c.__reactInternalSnapshotBeforeUpdate)}var g=i.updateQueue;g!==null&&Lh(i,g,c);break;case 3:var M=i.updateQueue;if(M!==null){if(o=null,i.child!==null)switch(i.child.tag){case 5:o=i.child.stateNode;break;case 1:o=i.child.stateNode}Lh(i,M,o)}break;case 5:var I=i.stateNode;if(o===null&&i.flags&4){o=I;var B=i.memoizedProps;switch(i.type){case"button":case"input":case"select":case"textarea":B.autoFocus&&o.focus();break;case"img":B.src&&(o.src=B.src)}}break;case 6:break;case 4:break;case 12:break;case 13:if(i.memoizedState===null){var ie=i.alternate;if(ie!==null){var _e=ie.memoizedState;if(_e!==null){var ye=_e.dehydrated;ye!==null&&_o(ye)}}}break;case 19:case 17:case 21:case 22:case 23:case 25:break;default:throw Error(t(163))}mn||i.flags&512&&gu(i)}catch(ge){Ht(i,i.return,ge)}}if(i===n){ze=null;break}if(o=i.sibling,o!==null){o.return=i.return,ze=o;break}ze=i.return}}function Up(n){for(;ze!==null;){var i=ze;if(i===n){ze=null;break}var o=i.sibling;if(o!==null){o.return=i.return,ze=o;break}ze=i.return}}function Fp(n){for(;ze!==null;){var i=ze;try{switch(i.tag){case 0:case 11:case 15:var o=i.return;try{Ka(4,i)}catch(B){Ht(i,o,B)}break;case 1:var c=i.stateNode;if(typeof c.componentDidMount=="function"){var d=i.return;try{c.componentDidMount()}catch(B){Ht(i,d,B)}}var g=i.return;try{gu(i)}catch(B){Ht(i,g,B)}break;case 5:var M=i.return;try{gu(i)}catch(B){Ht(i,M,B)}}}catch(B){Ht(i,i.return,B)}if(i===n){ze=null;break}var I=i.sibling;if(I!==null){I.return=i.return,ze=I;break}ze=i.return}}var Pv=Math.ceil,Za=D.ReactCurrentDispatcher,xu=D.ReactCurrentOwner,qn=D.ReactCurrentBatchConfig,_t=0,rn=null,jt=null,un=0,kn=0,Ps=nr(0),Kt=0,zo=null,kr=0,Qa=0,yu=0,Ho=null,An=null,Su=0,Ls=1/0,Di=null,Ja=!1,Mu=null,lr=null,el=!1,cr=null,tl=0,Vo=0,Eu=null,nl=-1,il=0;function yn(){return(_t&6)!==0?Ae():nl!==-1?nl:nl=Ae()}function ur(n){return(n.mode&1)===0?1:(_t&2)!==0&&un!==0?un&-un:hv.transition!==null?(il===0&&(il=us()),il):(n=wt,n!==0||(n=window.event,n=n===void 0?16:Od(n.type)),n)}function si(n,i,o,c){if(50<Vo)throw Vo=0,Eu=null,Error(t(185));qi(n,o,c),((_t&2)===0||n!==rn)&&(n===rn&&((_t&2)===0&&(Qa|=o),Kt===4&&fr(n,un)),Cn(n,c),o===1&&_t===0&&(i.mode&1)===0&&(Ls=Ae()+500,La&&rr()))}function Cn(n,i){var o=n.callbackNode;uo(n,i);var c=Lt(n,n===rn?un:0);if(c===0)o!==null&&j(o),n.callbackNode=null,n.callbackPriority=0;else if(i=c&-c,n.callbackPriority!==i){if(o!=null&&j(o),i===1)n.tag===0?dv(kp.bind(null,n)):Mh(kp.bind(null,n)),lv(function(){(_t&6)===0&&rr()}),o=null;else{switch(bd(c)){case 1:o=Ke;break;case 4:o=Qe;break;case 16:o=je;break;case 536870912:o=Ct;break;default:o=je}o=Xp(o,Op.bind(null,n))}n.callbackPriority=i,n.callbackNode=o}}function Op(n,i){if(nl=-1,il=0,(_t&6)!==0)throw Error(t(327));var o=n.callbackNode;if(Ds()&&n.callbackNode!==o)return null;var c=Lt(n,n===rn?un:0);if(c===0)return null;if((c&30)!==0||(c&n.expiredLanes)!==0||i)i=rl(n,c);else{i=c;var d=_t;_t|=2;var g=zp();(rn!==n||un!==i)&&(Di=null,Ls=Ae()+500,zr(n,i));do try{Nv();break}catch(I){Bp(n,I)}while(!0);Hc(),Za.current=g,_t=d,jt!==null?i=0:(rn=null,un=0,i=Kt)}if(i!==0){if(i===2&&(d=fn(n),d!==0&&(c=d,i=Tu(n,d))),i===1)throw o=zo,zr(n,0),fr(n,c),Cn(n,Ae()),o;if(i===6)fr(n,c);else{if(d=n.current.alternate,(c&30)===0&&!Lv(d)&&(i=rl(n,c),i===2&&(g=fn(n),g!==0&&(c=g,i=Tu(n,g))),i===1))throw o=zo,zr(n,0),fr(n,c),Cn(n,Ae()),o;switch(n.finishedWork=d,n.finishedLanes=c,i){case 0:case 1:throw Error(t(345));case 2:Hr(n,An,Di);break;case 3:if(fr(n,c),(c&130023424)===c&&(i=Su+500-Ae(),10<i)){if(Lt(n,0)!==0)break;if(d=n.suspendedLanes,(d&c)!==c){yn(),n.pingedLanes|=n.suspendedLanes&d;break}n.timeoutHandle=Pc(Hr.bind(null,n,An,Di),i);break}Hr(n,An,Di);break;case 4:if(fr(n,c),(c&4194240)===c)break;for(i=n.eventTimes,d=-1;0<c;){var M=31-ke(c);g=1<<M,M=i[M],M>d&&(d=M),c&=~g}if(c=d,c=Ae()-c,c=(120>c?120:480>c?480:1080>c?1080:1920>c?1920:3e3>c?3e3:4320>c?4320:1960*Pv(c/1960))-c,10<c){n.timeoutHandle=Pc(Hr.bind(null,n,An,Di),c);break}Hr(n,An,Di);break;case 5:Hr(n,An,Di);break;default:throw Error(t(329))}}}return Cn(n,Ae()),n.callbackNode===o?Op.bind(null,n):null}function Tu(n,i){var o=Ho;return n.current.memoizedState.isDehydrated&&(zr(n,i).flags|=256),n=rl(n,i),n!==2&&(i=An,An=o,i!==null&&wu(i)),n}function wu(n){An===null?An=n:An.push.apply(An,n)}function Lv(n){for(var i=n;;){if(i.flags&16384){var o=i.updateQueue;if(o!==null&&(o=o.stores,o!==null))for(var c=0;c<o.length;c++){var d=o[c],g=d.getSnapshot;d=d.value;try{if(!ei(g(),d))return!1}catch{return!1}}}if(o=i.child,i.subtreeFlags&16384&&o!==null)o.return=i,i=o;else{if(i===n)break;for(;i.sibling===null;){if(i.return===null||i.return===n)return!0;i=i.return}i.sibling.return=i.return,i=i.sibling}}return!0}function fr(n,i){for(i&=~yu,i&=~Qa,n.suspendedLanes|=i,n.pingedLanes&=~i,n=n.expirationTimes;0<i;){var o=31-ke(i),c=1<<o;n[o]=-1,i&=~c}}function kp(n){if((_t&6)!==0)throw Error(t(327));Ds();var i=Lt(n,0);if((i&1)===0)return Cn(n,Ae()),null;var o=rl(n,i);if(n.tag!==0&&o===2){var c=fn(n);c!==0&&(i=c,o=Tu(n,c))}if(o===1)throw o=zo,zr(n,0),fr(n,i),Cn(n,Ae()),o;if(o===6)throw Error(t(345));return n.finishedWork=n.current.alternate,n.finishedLanes=i,Hr(n,An,Di),Cn(n,Ae()),null}function Au(n,i){var o=_t;_t|=1;try{return n(i)}finally{_t=o,_t===0&&(Ls=Ae()+500,La&&rr())}}function Br(n){cr!==null&&cr.tag===0&&(_t&6)===0&&Ds();var i=_t;_t|=1;var o=qn.transition,c=wt;try{if(qn.transition=null,wt=1,n)return n()}finally{wt=c,qn.transition=o,_t=i,(_t&6)===0&&rr()}}function Cu(){kn=Ps.current,It(Ps)}function zr(n,i){n.finishedWork=null,n.finishedLanes=0;var o=n.timeoutHandle;if(o!==-1&&(n.timeoutHandle=-1,av(o)),jt!==null)for(o=jt.return;o!==null;){var c=o;switch(Fc(c),c.tag){case 1:c=c.type.childContextTypes,c!=null&&ba();break;case 3:Cs(),It(En),It(dn),Zc();break;case 5:$c(c);break;case 4:Cs();break;case 13:It(Bt);break;case 19:It(Bt);break;case 10:Vc(c.type._context);break;case 22:case 23:Cu()}o=o.return}if(rn=n,jt=n=dr(n.current,null),un=kn=i,Kt=0,zo=null,yu=Qa=kr=0,An=Ho=null,Ur!==null){for(i=0;i<Ur.length;i++)if(o=Ur[i],c=o.interleaved,c!==null){o.interleaved=null;var d=c.next,g=o.pending;if(g!==null){var M=g.next;g.next=d,c.next=M}o.pending=c}Ur=null}return n}function Bp(n,i){do{var o=jt;try{if(Hc(),Va.current=Xa,Ga){for(var c=zt.memoizedState;c!==null;){var d=c.queue;d!==null&&(d.pending=null),c=c.next}Ga=!1}if(Or=0,nn=$t=zt=null,Io=!1,Uo=0,xu.current=null,o===null||o.return===null){Kt=1,zo=i,jt=null;break}e:{var g=n,M=o.return,I=o,B=i;if(i=un,I.flags|=32768,B!==null&&typeof B=="object"&&typeof B.then=="function"){var ie=B,_e=I,ye=_e.tag;if((_e.mode&1)===0&&(ye===0||ye===11||ye===15)){var ge=_e.alternate;ge?(_e.updateQueue=ge.updateQueue,_e.memoizedState=ge.memoizedState,_e.lanes=ge.lanes):(_e.updateQueue=null,_e.memoizedState=null)}var Ue=up(M);if(Ue!==null){Ue.flags&=-257,fp(Ue,M,I,g,i),Ue.mode&1&&cp(g,ie,i),i=Ue,B=ie;var He=i.updateQueue;if(He===null){var Ve=new Set;Ve.add(B),i.updateQueue=Ve}else He.add(B);break e}else{if((i&1)===0){cp(g,ie,i),Ru();break e}B=Error(t(426))}}else if(Ot&&I.mode&1){var Wt=up(M);if(Wt!==null){(Wt.flags&65536)===0&&(Wt.flags|=256),fp(Wt,M,I,g,i),Bc(Rs(B,I));break e}}g=B=Rs(B,I),Kt!==4&&(Kt=2),Ho===null?Ho=[g]:Ho.push(g),g=M;do{switch(g.tag){case 3:g.flags|=65536,i&=-i,g.lanes|=i;var q=ap(g,B,i);Ph(g,q);break e;case 1:I=B;var H=g.type,$=g.stateNode;if((g.flags&128)===0&&(typeof H.getDerivedStateFromError=="function"||$!==null&&typeof $.componentDidCatch=="function"&&(lr===null||!lr.has($)))){g.flags|=65536,i&=-i,g.lanes|=i;var Te=lp(g,I,i);Ph(g,Te);break e}}g=g.return}while(g!==null)}Vp(o)}catch(Xe){i=Xe,jt===o&&o!==null&&(jt=o=o.return);continue}break}while(!0)}function zp(){var n=Za.current;return Za.current=Xa,n===null?Xa:n}function Ru(){(Kt===0||Kt===3||Kt===2)&&(Kt=4),rn===null||(kr&268435455)===0&&(Qa&268435455)===0||fr(rn,un)}function rl(n,i){var o=_t;_t|=2;var c=zp();(rn!==n||un!==i)&&(Di=null,zr(n,i));do try{Dv();break}catch(d){Bp(n,d)}while(!0);if(Hc(),_t=o,Za.current=c,jt!==null)throw Error(t(261));return rn=null,un=0,Kt}function Dv(){for(;jt!==null;)Hp(jt)}function Nv(){for(;jt!==null&&!we();)Hp(jt)}function Hp(n){var i=jp(n.alternate,n,kn);n.memoizedProps=n.pendingProps,i===null?Vp(n):jt=i,xu.current=null}function Vp(n){var i=n;do{var o=i.alternate;if(n=i.return,(i.flags&32768)===0){if(o=wv(o,i,kn),o!==null){jt=o;return}}else{if(o=Av(o,i),o!==null){o.flags&=32767,jt=o;return}if(n!==null)n.flags|=32768,n.subtreeFlags=0,n.deletions=null;else{Kt=6,jt=null;return}}if(i=i.sibling,i!==null){jt=i;return}jt=i=n}while(i!==null);Kt===0&&(Kt=5)}function Hr(n,i,o){var c=wt,d=qn.transition;try{qn.transition=null,wt=1,Iv(n,i,o,c)}finally{qn.transition=d,wt=c}return null}function Iv(n,i,o,c){do Ds();while(cr!==null);if((_t&6)!==0)throw Error(t(327));o=n.finishedWork;var d=n.finishedLanes;if(o===null)return null;if(n.finishedWork=null,n.finishedLanes=0,o===n.current)throw Error(t(177));n.callbackNode=null,n.callbackPriority=0;var g=o.lanes|o.childLanes;if(p_(n,g),n===rn&&(jt=rn=null,un=0),(o.subtreeFlags&2064)===0&&(o.flags&2064)===0||el||(el=!0,Xp(je,function(){return Ds(),null})),g=(o.flags&15990)!==0,(o.subtreeFlags&15990)!==0||g){g=qn.transition,qn.transition=null;var M=wt;wt=1;var I=_t;_t|=4,xu.current=null,Rv(n,o),Dp(o,n),ev(Rc),ma=!!Cc,Rc=Cc=null,n.current=o,bv(o),De(),_t=I,wt=M,qn.transition=g}else n.current=o;if(el&&(el=!1,cr=n,tl=d),g=n.pendingLanes,g===0&&(lr=null),xt(o.stateNode),Cn(n,Ae()),i!==null)for(c=n.onRecoverableError,o=0;o<i.length;o++)d=i[o],c(d.value,{componentStack:d.stack,digest:d.digest});if(Ja)throw Ja=!1,n=Mu,Mu=null,n;return(tl&1)!==0&&n.tag!==0&&Ds(),g=n.pendingLanes,(g&1)!==0?n===Eu?Vo++:(Vo=0,Eu=n):Vo=0,rr(),null}function Ds(){if(cr!==null){var n=bd(tl),i=qn.transition,o=wt;try{if(qn.transition=null,wt=16>n?16:n,cr===null)var c=!1;else{if(n=cr,cr=null,tl=0,(_t&6)!==0)throw Error(t(331));var d=_t;for(_t|=4,ze=n.current;ze!==null;){var g=ze,M=g.child;if((ze.flags&16)!==0){var I=g.deletions;if(I!==null){for(var B=0;B<I.length;B++){var ie=I[B];for(ze=ie;ze!==null;){var _e=ze;switch(_e.tag){case 0:case 11:case 15:Bo(8,_e,g)}var ye=_e.child;if(ye!==null)ye.return=_e,ze=ye;else for(;ze!==null;){_e=ze;var ge=_e.sibling,Ue=_e.return;if(Cp(_e),_e===ie){ze=null;break}if(ge!==null){ge.return=Ue,ze=ge;break}ze=Ue}}}var He=g.alternate;if(He!==null){var Ve=He.child;if(Ve!==null){He.child=null;do{var Wt=Ve.sibling;Ve.sibling=null,Ve=Wt}while(Ve!==null)}}ze=g}}if((g.subtreeFlags&2064)!==0&&M!==null)M.return=g,ze=M;else e:for(;ze!==null;){if(g=ze,(g.flags&2048)!==0)switch(g.tag){case 0:case 11:case 15:Bo(9,g,g.return)}var q=g.sibling;if(q!==null){q.return=g.return,ze=q;break e}ze=g.return}}var H=n.current;for(ze=H;ze!==null;){M=ze;var $=M.child;if((M.subtreeFlags&2064)!==0&&$!==null)$.return=M,ze=$;else e:for(M=H;ze!==null;){if(I=ze,(I.flags&2048)!==0)try{switch(I.tag){case 0:case 11:case 15:Ka(9,I)}}catch(Xe){Ht(I,I.return,Xe)}if(I===M){ze=null;break e}var Te=I.sibling;if(Te!==null){Te.return=I.return,ze=Te;break e}ze=I.return}}if(_t=d,rr(),Ft&&typeof Ft.onPostCommitFiberRoot=="function")try{Ft.onPostCommitFiberRoot(bt,n)}catch{}c=!0}return c}finally{wt=o,qn.transition=i}}return!1}function Gp(n,i,o){i=Rs(o,i),i=ap(n,i,1),n=or(n,i,1),i=yn(),n!==null&&(qi(n,1,i),Cn(n,i))}function Ht(n,i,o){if(n.tag===3)Gp(n,n,o);else for(;i!==null;){if(i.tag===3){Gp(i,n,o);break}else if(i.tag===1){var c=i.stateNode;if(typeof i.type.getDerivedStateFromError=="function"||typeof c.componentDidCatch=="function"&&(lr===null||!lr.has(c))){n=Rs(o,n),n=lp(i,n,1),i=or(i,n,1),n=yn(),i!==null&&(qi(i,1,n),Cn(i,n));break}}i=i.return}}function Uv(n,i,o){var c=n.pingCache;c!==null&&c.delete(i),i=yn(),n.pingedLanes|=n.suspendedLanes&o,rn===n&&(un&o)===o&&(Kt===4||Kt===3&&(un&130023424)===un&&500>Ae()-Su?zr(n,0):yu|=o),Cn(n,i)}function Wp(n,i){i===0&&((n.mode&1)===0?i=1:(i=tn,tn<<=1,(tn&130023424)===0&&(tn=4194304)));var o=yn();n=bi(n,i),n!==null&&(qi(n,i,o),Cn(n,o))}function Fv(n){var i=n.memoizedState,o=0;i!==null&&(o=i.retryLane),Wp(n,o)}function Ov(n,i){var o=0;switch(n.tag){case 13:var c=n.stateNode,d=n.memoizedState;d!==null&&(o=d.retryLane);break;case 19:c=n.stateNode;break;default:throw Error(t(314))}c!==null&&c.delete(i),Wp(n,o)}var jp;jp=function(n,i,o){if(n!==null)if(n.memoizedProps!==i.pendingProps||En.current)wn=!0;else{if((n.lanes&o)===0&&(i.flags&128)===0)return wn=!1,Tv(n,i,o);wn=(n.flags&131072)!==0}else wn=!1,Ot&&(i.flags&1048576)!==0&&Eh(i,Na,i.index);switch(i.lanes=0,i.tag){case 2:var c=i.type;qa(n,i),n=i.pendingProps;var d=ys(i,dn.current);ws(i,o),d=eu(null,i,c,n,d,o);var g=tu();return i.flags|=1,typeof d=="object"&&d!==null&&typeof d.render=="function"&&d.$$typeof===void 0?(i.tag=1,i.memoizedState=null,i.updateQueue=null,Tn(c)?(g=!0,Pa(i)):g=!1,i.memoizedState=d.state!==null&&d.state!==void 0?d.state:null,jc(i),d.updater=Ba,i.stateNode=d,d._reactInternals=i,Yc(i,c,n,o),i=cu(null,i,c,!0,g,o)):(i.tag=0,Ot&&g&&Uc(i),xn(null,i,d,o),i=i.child),i;case 16:c=i.elementType;e:{switch(qa(n,i),n=i.pendingProps,d=c._init,c=d(c._payload),i.type=c,d=i.tag=Bv(c),n=ni(c,n),d){case 0:i=lu(null,i,c,n,o);break e;case 1:i=_p(null,i,c,n,o);break e;case 11:i=dp(null,i,c,n,o);break e;case 14:i=hp(null,i,c,ni(c.type,n),o);break e}throw Error(t(306,c,""))}return i;case 0:return c=i.type,d=i.pendingProps,d=i.elementType===c?d:ni(c,d),lu(n,i,c,d,o);case 1:return c=i.type,d=i.pendingProps,d=i.elementType===c?d:ni(c,d),_p(n,i,c,d,o);case 3:e:{if(vp(i),n===null)throw Error(t(387));c=i.pendingProps,g=i.memoizedState,d=g.element,bh(n,i),ka(i,c,null,o);var M=i.memoizedState;if(c=M.element,g.isDehydrated)if(g={element:c,isDehydrated:!1,cache:M.cache,pendingSuspenseBoundaries:M.pendingSuspenseBoundaries,transitions:M.transitions},i.updateQueue.baseState=g,i.memoizedState=g,i.flags&256){d=Rs(Error(t(423)),i),i=xp(n,i,c,o,d);break e}else if(c!==d){d=Rs(Error(t(424)),i),i=xp(n,i,c,o,d);break e}else for(On=tr(i.stateNode.containerInfo.firstChild),Fn=i,Ot=!0,ti=null,o=kh(i,null,c,o),i.child=o;o;)o.flags=o.flags&-3|4096,o=o.sibling;else{if(Es(),c===d){i=Li(n,i,o);break e}xn(n,i,c,o)}i=i.child}return i;case 5:return Bh(i),n===null&&kc(i),c=i.type,d=i.pendingProps,g=n!==null?n.memoizedProps:null,M=d.children,bc(c,d)?M=null:g!==null&&bc(c,g)&&(i.flags|=32),gp(n,i),xn(n,i,M,o),i.child;case 6:return n===null&&kc(i),null;case 13:return yp(n,i,o);case 4:return qc(i,i.stateNode.containerInfo),c=i.pendingProps,n===null?i.child=As(i,null,c,o):xn(n,i,c,o),i.child;case 11:return c=i.type,d=i.pendingProps,d=i.elementType===c?d:ni(c,d),dp(n,i,c,d,o);case 7:return xn(n,i,i.pendingProps,o),i.child;case 8:return xn(n,i,i.pendingProps.children,o),i.child;case 12:return xn(n,i,i.pendingProps.children,o),i.child;case 10:e:{if(c=i.type._context,d=i.pendingProps,g=i.memoizedProps,M=d.value,Dt(Ua,c._currentValue),c._currentValue=M,g!==null)if(ei(g.value,M)){if(g.children===d.children&&!En.current){i=Li(n,i,o);break e}}else for(g=i.child,g!==null&&(g.return=i);g!==null;){var I=g.dependencies;if(I!==null){M=g.child;for(var B=I.firstContext;B!==null;){if(B.context===c){if(g.tag===1){B=Pi(-1,o&-o),B.tag=2;var ie=g.updateQueue;if(ie!==null){ie=ie.shared;var _e=ie.pending;_e===null?B.next=B:(B.next=_e.next,_e.next=B),ie.pending=B}}g.lanes|=o,B=g.alternate,B!==null&&(B.lanes|=o),Gc(g.return,o,i),I.lanes|=o;break}B=B.next}}else if(g.tag===10)M=g.type===i.type?null:g.child;else if(g.tag===18){if(M=g.return,M===null)throw Error(t(341));M.lanes|=o,I=M.alternate,I!==null&&(I.lanes|=o),Gc(M,o,i),M=g.sibling}else M=g.child;if(M!==null)M.return=g;else for(M=g;M!==null;){if(M===i){M=null;break}if(g=M.sibling,g!==null){g.return=M.return,M=g;break}M=M.return}g=M}xn(n,i,d.children,o),i=i.child}return i;case 9:return d=i.type,c=i.pendingProps.children,ws(i,o),d=Xn(d),c=c(d),i.flags|=1,xn(n,i,c,o),i.child;case 14:return c=i.type,d=ni(c,i.pendingProps),d=ni(c.type,d),hp(n,i,c,d,o);case 15:return pp(n,i,i.type,i.pendingProps,o);case 17:return c=i.type,d=i.pendingProps,d=i.elementType===c?d:ni(c,d),qa(n,i),i.tag=1,Tn(c)?(n=!0,Pa(i)):n=!1,ws(i,o),Ih(i,c,d),Yc(i,c,d,o),cu(null,i,c,!0,n,o);case 19:return Mp(n,i,o);case 22:return mp(n,i,o)}throw Error(t(156,i.tag))};function Xp(n,i){return ne(n,i)}function kv(n,i,o,c){this.tag=n,this.key=o,this.sibling=this.child=this.return=this.stateNode=this.type=this.elementType=null,this.index=0,this.ref=null,this.pendingProps=i,this.dependencies=this.memoizedState=this.updateQueue=this.memoizedProps=null,this.mode=c,this.subtreeFlags=this.flags=0,this.deletions=null,this.childLanes=this.lanes=0,this.alternate=null}function $n(n,i,o,c){return new kv(n,i,o,c)}function bu(n){return n=n.prototype,!(!n||!n.isReactComponent)}function Bv(n){if(typeof n=="function")return bu(n)?1:0;if(n!=null){if(n=n.$$typeof,n===re)return 11;if(n===ue)return 14}return 2}function dr(n,i){var o=n.alternate;return o===null?(o=$n(n.tag,i,n.key,n.mode),o.elementType=n.elementType,o.type=n.type,o.stateNode=n.stateNode,o.alternate=n,n.alternate=o):(o.pendingProps=i,o.type=n.type,o.flags=0,o.subtreeFlags=0,o.deletions=null),o.flags=n.flags&14680064,o.childLanes=n.childLanes,o.lanes=n.lanes,o.child=n.child,o.memoizedProps=n.memoizedProps,o.memoizedState=n.memoizedState,o.updateQueue=n.updateQueue,i=n.dependencies,o.dependencies=i===null?null:{lanes:i.lanes,firstContext:i.firstContext},o.sibling=n.sibling,o.index=n.index,o.ref=n.ref,o}function sl(n,i,o,c,d,g){var M=2;if(c=n,typeof n=="function")bu(n)&&(M=1);else if(typeof n=="string")M=5;else e:switch(n){case U:return Vr(o.children,d,g,i);case Y:M=8,d|=8;break;case ce:return n=$n(12,o,i,d|2),n.elementType=ce,n.lanes=g,n;case ee:return n=$n(13,o,i,d),n.elementType=ee,n.lanes=g,n;case ae:return n=$n(19,o,i,d),n.elementType=ae,n.lanes=g,n;case le:return ol(o,d,g,i);default:if(typeof n=="object"&&n!==null)switch(n.$$typeof){case E:M=10;break e;case C:M=9;break e;case re:M=11;break e;case ue:M=14;break e;case Z:M=16,c=null;break e}throw Error(t(130,n==null?n:typeof n,""))}return i=$n(M,o,i,d),i.elementType=n,i.type=c,i.lanes=g,i}function Vr(n,i,o,c){return n=$n(7,n,c,i),n.lanes=o,n}function ol(n,i,o,c){return n=$n(22,n,c,i),n.elementType=le,n.lanes=o,n.stateNode={isHidden:!1},n}function Pu(n,i,o){return n=$n(6,n,null,i),n.lanes=o,n}function Lu(n,i,o){return i=$n(4,n.children!==null?n.children:[],n.key,i),i.lanes=o,i.stateNode={containerInfo:n.containerInfo,pendingChildren:null,implementation:n.implementation},i}function zv(n,i,o,c,d){this.tag=i,this.containerInfo=n,this.finishedWork=this.pingCache=this.current=this.pendingChildren=null,this.timeoutHandle=-1,this.callbackNode=this.pendingContext=this.context=null,this.callbackPriority=0,this.eventTimes=fo(0),this.expirationTimes=fo(-1),this.entangledLanes=this.finishedLanes=this.mutableReadLanes=this.expiredLanes=this.pingedLanes=this.suspendedLanes=this.pendingLanes=0,this.entanglements=fo(0),this.identifierPrefix=c,this.onRecoverableError=d,this.mutableSourceEagerHydrationData=null}function Du(n,i,o,c,d,g,M,I,B){return n=new zv(n,i,o,I,B),i===1?(i=1,g===!0&&(i|=8)):i=0,g=$n(3,null,null,i),n.current=g,g.stateNode=n,g.memoizedState={element:c,isDehydrated:o,cache:null,transitions:null,pendingSuspenseBoundaries:null},jc(g),n}function Hv(n,i,o){var c=3<arguments.length&&arguments[3]!==void 0?arguments[3]:null;return{$$typeof:O,key:c==null?null:""+c,children:n,containerInfo:i,implementation:o}}function Yp(n){if(!n)return ir;n=n._reactInternals;e:{if(Ei(n)!==n||n.tag!==1)throw Error(t(170));var i=n;do{switch(i.tag){case 3:i=i.stateNode.context;break e;case 1:if(Tn(i.type)){i=i.stateNode.__reactInternalMemoizedMergedChildContext;break e}}i=i.return}while(i!==null);throw Error(t(171))}if(n.tag===1){var o=n.type;if(Tn(o))return yh(n,o,i)}return i}function qp(n,i,o,c,d,g,M,I,B){return n=Du(o,c,!0,n,d,g,M,I,B),n.context=Yp(null),o=n.current,c=yn(),d=ur(o),g=Pi(c,d),g.callback=i??null,or(o,g,d),n.current.lanes=d,qi(n,d,c),Cn(n,c),n}function al(n,i,o,c){var d=i.current,g=yn(),M=ur(d);return o=Yp(o),i.context===null?i.context=o:i.pendingContext=o,i=Pi(g,M),i.payload={element:n},c=c===void 0?null:c,c!==null&&(i.callback=c),n=or(d,i,M),n!==null&&(si(n,d,M,g),Oa(n,d,M)),M}function ll(n){return n=n.current,n.child?(n.child.tag===5,n.child.stateNode):null}function $p(n,i){if(n=n.memoizedState,n!==null&&n.dehydrated!==null){var o=n.retryLane;n.retryLane=o!==0&&o<i?o:i}}function Nu(n,i){$p(n,i),(n=n.alternate)&&$p(n,i)}function Vv(){return null}var Kp=typeof reportError=="function"?reportError:function(n){console.error(n)};function Iu(n){this._internalRoot=n}cl.prototype.render=Iu.prototype.render=function(n){var i=this._internalRoot;if(i===null)throw Error(t(409));al(n,i,null,null)},cl.prototype.unmount=Iu.prototype.unmount=function(){var n=this._internalRoot;if(n!==null){this._internalRoot=null;var i=n.containerInfo;Br(function(){al(null,n,null,null)}),i[wi]=null}};function cl(n){this._internalRoot=n}cl.prototype.unstable_scheduleHydration=function(n){if(n){var i=Dd();n={blockedOn:null,target:n,priority:i};for(var o=0;o<Qi.length&&i!==0&&i<Qi[o].priority;o++);Qi.splice(o,0,n),o===0&&Ud(n)}};function Uu(n){return!(!n||n.nodeType!==1&&n.nodeType!==9&&n.nodeType!==11)}function ul(n){return!(!n||n.nodeType!==1&&n.nodeType!==9&&n.nodeType!==11&&(n.nodeType!==8||n.nodeValue!==" react-mount-point-unstable "))}function Zp(){}function Gv(n,i,o,c,d){if(d){if(typeof c=="function"){var g=c;c=function(){var ie=ll(M);g.call(ie)}}var M=qp(i,c,n,0,null,!1,!1,"",Zp);return n._reactRootContainer=M,n[wi]=M.current,Ao(n.nodeType===8?n.parentNode:n),Br(),M}for(;d=n.lastChild;)n.removeChild(d);if(typeof c=="function"){var I=c;c=function(){var ie=ll(B);I.call(ie)}}var B=Du(n,0,!1,null,null,!1,!1,"",Zp);return n._reactRootContainer=B,n[wi]=B.current,Ao(n.nodeType===8?n.parentNode:n),Br(function(){al(i,B,o,c)}),B}function fl(n,i,o,c,d){var g=o._reactRootContainer;if(g){var M=g;if(typeof d=="function"){var I=d;d=function(){var B=ll(M);I.call(B)}}al(i,M,n,d)}else M=Gv(o,i,n,d,c);return ll(M)}Pd=function(n){switch(n.tag){case 3:var i=n.stateNode;if(i.current.memoizedState.isDehydrated){var o=Ti(i.pendingLanes);o!==0&&(sc(i,o|1),Cn(i,Ae()),(_t&6)===0&&(Ls=Ae()+500,rr()))}break;case 13:Br(function(){var c=bi(n,1);if(c!==null){var d=yn();si(c,n,1,d)}}),Nu(n,1)}},oc=function(n){if(n.tag===13){var i=bi(n,134217728);if(i!==null){var o=yn();si(i,n,134217728,o)}Nu(n,134217728)}},Ld=function(n){if(n.tag===13){var i=ur(n),o=bi(n,i);if(o!==null){var c=yn();si(o,n,i,c)}Nu(n,i)}},Dd=function(){return wt},Nd=function(n,i){var o=wt;try{return wt=n,i()}finally{wt=o}},oe=function(n,i,o){switch(i){case"input":if(Ze(n,o),i=o.name,o.type==="radio"&&i!=null){for(o=n;o.parentNode;)o=o.parentNode;for(o=o.querySelectorAll("input[name="+JSON.stringify(""+i)+'][type="radio"]'),i=0;i<o.length;i++){var c=o[i];if(c!==n&&c.form===n.form){var d=Ra(c);if(!d)throw Error(t(90));ct(c),Ze(c,d)}}}break;case"textarea":xe(n,o);break;case"select":i=o.value,i!=null&&A(n,!!o.multiple,i,!1)}},ln=Au,mt=Br;var Wv={usingClientEntryPoint:!1,Events:[bo,vs,Ra,ft,kt,Au]},Go={findFiberByHostInstance:Lr,bundleType:0,version:"18.2.0",rendererPackageName:"react-dom"},jv={bundleType:Go.bundleType,version:Go.version,rendererPackageName:Go.rendererPackageName,rendererConfig:Go.rendererConfig,overrideHookState:null,overrideHookStateDeletePath:null,overrideHookStateRenamePath:null,overrideProps:null,overridePropsDeletePath:null,overridePropsRenamePath:null,setErrorHandler:null,setSuspenseHandler:null,scheduleUpdate:null,currentDispatcherRef:D.ReactCurrentDispatcher,findHostInstanceByFiber:function(n){return n=W(n),n===null?null:n.stateNode},findFiberByHostInstance:Go.findFiberByHostInstance||Vv,findHostInstancesForRefresh:null,scheduleRefresh:null,scheduleRoot:null,setRefreshHandler:null,getCurrentFiber:null,reconcilerVersion:"18.2.0-next-9e3b772b8-20220608"};if(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__<"u"){var dl=__REACT_DEVTOOLS_GLOBAL_HOOK__;if(!dl.isDisabled&&dl.supportsFiber)try{bt=dl.inject(jv),Ft=dl}catch{}}return Rn.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED=Wv,Rn.createPortal=function(n,i){var o=2<arguments.length&&arguments[2]!==void 0?arguments[2]:null;if(!Uu(i))throw Error(t(200));return Hv(n,i,null,o)},Rn.createRoot=function(n,i){if(!Uu(n))throw Error(t(299));var o=!1,c="",d=Kp;return i!=null&&(i.unstable_strictMode===!0&&(o=!0),i.identifierPrefix!==void 0&&(c=i.identifierPrefix),i.onRecoverableError!==void 0&&(d=i.onRecoverableError)),i=Du(n,1,!1,null,null,o,!1,c,d),n[wi]=i.current,Ao(n.nodeType===8?n.parentNode:n),new Iu(i)},Rn.findDOMNode=function(n){if(n==null)return null;if(n.nodeType===1)return n;var i=n._reactInternals;if(i===void 0)throw typeof n.render=="function"?Error(t(188)):(n=Object.keys(n).join(","),Error(t(268,n)));return n=W(i),n=n===null?null:n.stateNode,n},Rn.flushSync=function(n){return Br(n)},Rn.hydrate=function(n,i,o){if(!ul(i))throw Error(t(200));return fl(null,n,i,!0,o)},Rn.hydrateRoot=function(n,i,o){if(!Uu(n))throw Error(t(405));var c=o!=null&&o.hydratedSources||null,d=!1,g="",M=Kp;if(o!=null&&(o.unstable_strictMode===!0&&(d=!0),o.identifierPrefix!==void 0&&(g=o.identifierPrefix),o.onRecoverableError!==void 0&&(M=o.onRecoverableError)),i=qp(i,null,n,1,o??null,d,!1,g,M),n[wi]=i.current,Ao(n),c)for(n=0;n<c.length;n++)o=c[n],d=o._getVersion,d=d(o._source),i.mutableSourceEagerHydrationData==null?i.mutableSourceEagerHydrationData=[o,d]:i.mutableSourceEagerHydrationData.push(o,d);return new cl(i)},Rn.render=function(n,i,o){if(!ul(i))throw Error(t(200));return fl(null,n,i,!1,o)},Rn.unmountComponentAtNode=function(n){if(!ul(n))throw Error(t(40));return n._reactRootContainer?(Br(function(){fl(null,null,n,!1,function(){n._reactRootContainer=null,n[wi]=null})}),!0):!1},Rn.unstable_batchedUpdates=Au,Rn.unstable_renderSubtreeIntoContainer=function(n,i,o,c){if(!ul(o))throw Error(t(200));if(n==null||n._reactInternals===void 0)throw Error(t(38));return fl(n,i,o,!1,c)},Rn.version="18.2.0-next-9e3b772b8-20220608",Rn}var sm;function _g(){if(sm)return ku.exports;sm=1;function r(){if(!(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__>"u"||typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE!="function"))try{__REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(r)}catch(e){console.error(e)}}return r(),ku.exports=n0(),ku.exports}var om;function i0(){if(om)return hl;om=1;var r=_g();return hl.createRoot=r.createRoot,hl.hydrateRoot=r.hydrateRoot,hl}var r0=i0();_g();function Jo(){return Jo=Object.assign?Object.assign.bind():function(r){for(var e=1;e<arguments.length;e++){var t=arguments[e];for(var s in t)Object.prototype.hasOwnProperty.call(t,s)&&(r[s]=t[s])}return r},Jo.apply(this,arguments)}var Mr;(function(r){r.Pop="POP",r.Push="PUSH",r.Replace="REPLACE"})(Mr||(Mr={}));const am="popstate";function s0(r){r===void 0&&(r={});function e(s,a){let{pathname:l,search:u,hash:f}=s.location;return Sf("",{pathname:l,search:u,hash:f},a.state&&a.state.usr||null,a.state&&a.state.key||"default")}function t(s,a){return typeof a=="string"?a:Xl(a)}return a0(e,t,null,r)}function Yt(r,e){if(r===!1||r===null||typeof r>"u")throw new Error(e)}function vg(r,e){if(!r){typeof console<"u"&&console.warn(e);try{throw new Error(e)}catch{}}}function o0(){return Math.random().toString(36).substr(2,8)}function lm(r,e){return{usr:r.state,key:r.key,idx:e}}function Sf(r,e,t,s){return t===void 0&&(t=null),Jo({pathname:typeof r=="string"?r:r.pathname,search:"",hash:""},typeof e=="string"?oo(e):e,{state:t,key:e&&e.key||s||o0()})}function Xl(r){let{pathname:e="/",search:t="",hash:s=""}=r;return t&&t!=="?"&&(e+=t.charAt(0)==="?"?t:"?"+t),s&&s!=="#"&&(e+=s.charAt(0)==="#"?s:"#"+s),e}function oo(r){let e={};if(r){let t=r.indexOf("#");t>=0&&(e.hash=r.substr(t),r=r.substr(0,t));let s=r.indexOf("?");s>=0&&(e.search=r.substr(s),r=r.substr(0,s)),r&&(e.pathname=r)}return e}function a0(r,e,t,s){s===void 0&&(s={});let{window:a=document.defaultView,v5Compat:l=!1}=s,u=a.history,f=Mr.Pop,h=null,p=m();p==null&&(p=0,u.replaceState(Jo({},u.state,{idx:p}),""));function m(){return(u.state||{idx:null}).idx}function _(){f=Mr.Pop;let v=m(),y=v==null?null:v-p;p=v,h&&h({action:f,location:w.location,delta:y})}function x(v,y){f=Mr.Push;let P=Sf(w.location,v,y);p=m()+1;let b=lm(P,p),D=w.createHref(P);try{u.pushState(b,"",D)}catch(V){if(V instanceof DOMException&&V.name==="DataCloneError")throw V;a.location.assign(D)}l&&h&&h({action:f,location:w.location,delta:1})}function S(v,y){f=Mr.Replace;let P=Sf(w.location,v,y);p=m();let b=lm(P,p),D=w.createHref(P);u.replaceState(b,"",D),l&&h&&h({action:f,location:w.location,delta:0})}function T(v){let y=a.location.origin!=="null"?a.location.origin:a.location.href,P=typeof v=="string"?v:Xl(v);return P=P.replace(/ $/,"%20"),Yt(y,"No window.location.(origin|href) available to create URL for href: "+P),new URL(P,y)}let w={get action(){return f},get location(){return r(a,u)},listen(v){if(h)throw new Error("A history only accepts one active listener");return a.addEventListener(am,_),h=v,()=>{a.removeEventListener(am,_),h=null}},createHref(v){return e(a,v)},createURL:T,encodeLocation(v){let y=T(v);return{pathname:y.pathname,search:y.search,hash:y.hash}},push:x,replace:S,go(v){return u.go(v)}};return w}var cm;(function(r){r.data="data",r.deferred="deferred",r.redirect="redirect",r.error="error"})(cm||(cm={}));function l0(r,e,t){t===void 0&&(t="/");let s=typeof e=="string"?oo(e):e,a=hd(s.pathname||"/",t);if(a==null)return null;let l=xg(r);c0(l);let u=null;for(let f=0;u==null&&f<l.length;++f){let h=S0(a);u=v0(l[f],h)}return u}function xg(r,e,t,s){e===void 0&&(e=[]),t===void 0&&(t=[]),s===void 0&&(s="");let a=(l,u,f)=>{let h={relativePath:f===void 0?l.path||"":f,caseSensitive:l.caseSensitive===!0,childrenIndex:u,route:l};h.relativePath.startsWith("/")&&(Yt(h.relativePath.startsWith(s),'Absolute route path "'+h.relativePath+'" nested under path '+('"'+s+'" is not valid. An absolute child route path ')+"must start with the combined path of all its parent routes."),h.relativePath=h.relativePath.slice(s.length));let p=Er([s,h.relativePath]),m=t.concat(h);l.children&&l.children.length>0&&(Yt(l.index!==!0,"Index routes must not have child routes. Please remove "+('all child routes from route path "'+p+'".')),xg(l.children,e,m,p)),!(l.path==null&&!l.index)&&e.push({path:p,score:g0(p,l.index),routesMeta:m})};return r.forEach((l,u)=>{var f;if(l.path===""||!((f=l.path)!=null&&f.includes("?")))a(l,u);else for(let h of yg(l.path))a(l,u,h)}),e}function yg(r){let e=r.split("/");if(e.length===0)return[];let[t,...s]=e,a=t.endsWith("?"),l=t.replace(/\?$/,"");if(s.length===0)return a?[l,""]:[l];let u=yg(s.join("/")),f=[];return f.push(...u.map(h=>h===""?l:[l,h].join("/"))),a&&f.push(...u),f.map(h=>r.startsWith("/")&&h===""?"/":h)}function c0(r){r.sort((e,t)=>e.score!==t.score?t.score-e.score:_0(e.routesMeta.map(s=>s.childrenIndex),t.routesMeta.map(s=>s.childrenIndex)))}const u0=/^:[\w-]+$/,f0=3,d0=2,h0=1,p0=10,m0=-2,um=r=>r==="*";function g0(r,e){let t=r.split("/"),s=t.length;return t.some(um)&&(s+=m0),e&&(s+=d0),t.filter(a=>!um(a)).reduce((a,l)=>a+(u0.test(l)?f0:l===""?h0:p0),s)}function _0(r,e){return r.length===e.length&&r.slice(0,-1).every((s,a)=>s===e[a])?r[r.length-1]-e[e.length-1]:0}function v0(r,e){let{routesMeta:t}=r,s={},a="/",l=[];for(let u=0;u<t.length;++u){let f=t[u],h=u===t.length-1,p=a==="/"?e:e.slice(a.length)||"/",m=x0({path:f.relativePath,caseSensitive:f.caseSensitive,end:h},p);if(!m)return null;Object.assign(s,m.params);let _=f.route;l.push({params:s,pathname:Er([a,m.pathname]),pathnameBase:w0(Er([a,m.pathnameBase])),route:_}),m.pathnameBase!=="/"&&(a=Er([a,m.pathnameBase]))}return l}function x0(r,e){typeof r=="string"&&(r={path:r,caseSensitive:!1,end:!0});let[t,s]=y0(r.path,r.caseSensitive,r.end),a=e.match(t);if(!a)return null;let l=a[0],u=l.replace(/(.)\/+$/,"$1"),f=a.slice(1);return{params:s.reduce((p,m,_)=>{let{paramName:x,isOptional:S}=m;if(x==="*"){let w=f[_]||"";u=l.slice(0,l.length-w.length).replace(/(.)\/+$/,"$1")}const T=f[_];return S&&!T?p[x]=void 0:p[x]=(T||"").replace(/%2F/g,"/"),p},{}),pathname:l,pathnameBase:u,pattern:r}}function y0(r,e,t){e===void 0&&(e=!1),t===void 0&&(t=!0),vg(r==="*"||!r.endsWith("*")||r.endsWith("/*"),'Route path "'+r+'" will be treated as if it were '+('"'+r.replace(/\*$/,"/*")+'" because the `*` character must ')+"always follow a `/` in the pattern. To get rid of this warning, "+('please change the route path to "'+r.replace(/\*$/,"/*")+'".'));let s=[],a="^"+r.replace(/\/*\*?$/,"").replace(/^\/*/,"/").replace(/[\\.*+^${}|()[\]]/g,"\\$&").replace(/\/:([\w-]+)(\?)?/g,(u,f,h)=>(s.push({paramName:f,isOptional:h!=null}),h?"/?([^\\/]+)?":"/([^\\/]+)"));return r.endsWith("*")?(s.push({paramName:"*"}),a+=r==="*"||r==="/*"?"(.*)$":"(?:\\/(.+)|\\/*)$"):t?a+="\\/*$":r!==""&&r!=="/"&&(a+="(?:(?=\\/|$))"),[new RegExp(a,e?void 0:"i"),s]}function S0(r){try{return r.split("/").map(e=>decodeURIComponent(e).replace(/\//g,"%2F")).join("/")}catch(e){return vg(!1,'The URL path "'+r+'" could not be decoded because it is is a malformed URL segment. This is probably due to a bad percent '+("encoding ("+e+").")),r}}function hd(r,e){if(e==="/")return r;if(!r.toLowerCase().startsWith(e.toLowerCase()))return null;let t=e.endsWith("/")?e.length-1:e.length,s=r.charAt(t);return s&&s!=="/"?null:r.slice(t)||"/"}function M0(r,e){e===void 0&&(e="/");let{pathname:t,search:s="",hash:a=""}=typeof r=="string"?oo(r):r;return{pathname:t?t.startsWith("/")?t:E0(t,e):e,search:A0(s),hash:C0(a)}}function E0(r,e){let t=e.replace(/\/+$/,"").split("/");return r.split("/").forEach(a=>{a===".."?t.length>1&&t.pop():a!=="."&&t.push(a)}),t.length>1?t.join("/"):"/"}function Hu(r,e,t,s){return"Cannot include a '"+r+"' character in a manually specified "+("`to."+e+"` field ["+JSON.stringify(s)+"].  Please separate it out to the ")+("`to."+t+"` field. Alternatively you may provide the full path as ")+'a string in <Link to="..."> and the router will parse it for you.'}function T0(r){return r.filter((e,t)=>t===0||e.route.path&&e.route.path.length>0)}function Sg(r,e){let t=T0(r);return e?t.map((s,a)=>a===r.length-1?s.pathname:s.pathnameBase):t.map(s=>s.pathnameBase)}function Mg(r,e,t,s){s===void 0&&(s=!1);let a;typeof r=="string"?a=oo(r):(a=Jo({},r),Yt(!a.pathname||!a.pathname.includes("?"),Hu("?","pathname","search",a)),Yt(!a.pathname||!a.pathname.includes("#"),Hu("#","pathname","hash",a)),Yt(!a.search||!a.search.includes("#"),Hu("#","search","hash",a)));let l=r===""||a.pathname==="",u=l?"/":a.pathname,f;if(u==null)f=t;else{let _=e.length-1;if(!s&&u.startsWith("..")){let x=u.split("/");for(;x[0]==="..";)x.shift(),_-=1;a.pathname=x.join("/")}f=_>=0?e[_]:"/"}let h=M0(a,f),p=u&&u!=="/"&&u.endsWith("/"),m=(l||u===".")&&t.endsWith("/");return!h.pathname.endsWith("/")&&(p||m)&&(h.pathname+="/"),h}const Er=r=>r.join("/").replace(/\/\/+/g,"/"),w0=r=>r.replace(/\/+$/,"").replace(/^\/*/,"/"),A0=r=>!r||r==="?"?"":r.startsWith("?")?r:"?"+r,C0=r=>!r||r==="#"?"":r.startsWith("#")?r:"#"+r;function R0(r){return r!=null&&typeof r.status=="number"&&typeof r.statusText=="string"&&typeof r.internal=="boolean"&&"data"in r}const Eg=["post","put","patch","delete"];new Set(Eg);const b0=["get",...Eg];new Set(b0);function ea(){return ea=Object.assign?Object.assign.bind():function(r){for(var e=1;e<arguments.length;e++){var t=arguments[e];for(var s in t)Object.prototype.hasOwnProperty.call(t,s)&&(r[s]=t[s])}return r},ea.apply(this,arguments)}const pd=he.createContext(null),P0=he.createContext(null),ss=he.createContext(null),Ql=he.createContext(null),Rr=he.createContext({outlet:null,matches:[],isDataRoute:!1}),Tg=he.createContext(null);function L0(r,e){let{relative:t}=e===void 0?{}:e;ia()||Yt(!1);let{basename:s,navigator:a}=he.useContext(ss),{hash:l,pathname:u,search:f}=Ag(r,{relative:t}),h=u;return s!=="/"&&(h=u==="/"?s:Er([s,u])),a.createHref({pathname:h,search:f,hash:l})}function ia(){return he.useContext(Ql)!=null}function ra(){return ia()||Yt(!1),he.useContext(Ql).location}function wg(r){he.useContext(ss).static||he.useLayoutEffect(r)}function md(){let{isDataRoute:r}=he.useContext(Rr);return r?W0():D0()}function D0(){ia()||Yt(!1);let r=he.useContext(pd),{basename:e,future:t,navigator:s}=he.useContext(ss),{matches:a}=he.useContext(Rr),{pathname:l}=ra(),u=JSON.stringify(Sg(a,t.v7_relativeSplatPath)),f=he.useRef(!1);return wg(()=>{f.current=!0}),he.useCallback(function(p,m){if(m===void 0&&(m={}),!f.current)return;if(typeof p=="number"){s.go(p);return}let _=Mg(p,JSON.parse(u),l,m.relative==="path");r==null&&e!=="/"&&(_.pathname=_.pathname==="/"?e:Er([e,_.pathname])),(m.replace?s.replace:s.push)(_,m.state,m)},[e,s,u,l,r])}function Jl(){let{matches:r}=he.useContext(Rr),e=r[r.length-1];return e?e.params:{}}function Ag(r,e){let{relative:t}=e===void 0?{}:e,{future:s}=he.useContext(ss),{matches:a}=he.useContext(Rr),{pathname:l}=ra(),u=JSON.stringify(Sg(a,s.v7_relativeSplatPath));return he.useMemo(()=>Mg(r,JSON.parse(u),l,t==="path"),[r,u,l,t])}function N0(r,e){return I0(r,e)}function I0(r,e,t,s){ia()||Yt(!1);let{navigator:a}=he.useContext(ss),{matches:l}=he.useContext(Rr),u=l[l.length-1],f=u?u.params:{};u&&u.pathname;let h=u?u.pathnameBase:"/";u&&u.route;let p=ra(),m;if(e){var _;let v=typeof e=="string"?oo(e):e;h==="/"||(_=v.pathname)!=null&&_.startsWith(h)||Yt(!1),m=v}else m=p;let x=m.pathname||"/",S=x;if(h!=="/"){let v=h.replace(/^\//,"").split("/");S="/"+x.replace(/^\//,"").split("/").slice(v.length).join("/")}let T=l0(r,{pathname:S}),w=B0(T&&T.map(v=>Object.assign({},v,{params:Object.assign({},f,v.params),pathname:Er([h,a.encodeLocation?a.encodeLocation(v.pathname).pathname:v.pathname]),pathnameBase:v.pathnameBase==="/"?h:Er([h,a.encodeLocation?a.encodeLocation(v.pathnameBase).pathname:v.pathnameBase])})),l,t,s);return e&&w?he.createElement(Ql.Provider,{value:{location:ea({pathname:"/",search:"",hash:"",state:null,key:"default"},m),navigationType:Mr.Pop}},w):w}function U0(){let r=G0(),e=R0(r)?r.status+" "+r.statusText:r instanceof Error?r.message:JSON.stringify(r),t=r instanceof Error?r.stack:null,a={padding:"0.5rem",backgroundColor:"rgba(200,200,200, 0.5)"};return he.createElement(he.Fragment,null,he.createElement("h2",null,"Unexpected Application Error!"),he.createElement("h3",{style:{fontStyle:"italic"}},e),t?he.createElement("pre",{style:a},t):null,null)}const F0=he.createElement(U0,null);class O0 extends he.Component{constructor(e){super(e),this.state={location:e.location,revalidation:e.revalidation,error:e.error}}static getDerivedStateFromError(e){return{error:e}}static getDerivedStateFromProps(e,t){return t.location!==e.location||t.revalidation!=="idle"&&e.revalidation==="idle"?{error:e.error,location:e.location,revalidation:e.revalidation}:{error:e.error!==void 0?e.error:t.error,location:t.location,revalidation:e.revalidation||t.revalidation}}componentDidCatch(e,t){console.error("React Router caught the following error during render",e,t)}render(){return this.state.error!==void 0?he.createElement(Rr.Provider,{value:this.props.routeContext},he.createElement(Tg.Provider,{value:this.state.error,children:this.props.component})):this.props.children}}function k0(r){let{routeContext:e,match:t,children:s}=r,a=he.useContext(pd);return a&&a.static&&a.staticContext&&(t.route.errorElement||t.route.ErrorBoundary)&&(a.staticContext._deepestRenderedBoundaryId=t.route.id),he.createElement(Rr.Provider,{value:e},s)}function B0(r,e,t,s){var a;if(e===void 0&&(e=[]),t===void 0&&(t=null),s===void 0&&(s=null),r==null){var l;if((l=t)!=null&&l.errors)r=t.matches;else return null}let u=r,f=(a=t)==null?void 0:a.errors;if(f!=null){let m=u.findIndex(_=>_.route.id&&f?.[_.route.id]);m>=0||Yt(!1),u=u.slice(0,Math.min(u.length,m+1))}let h=!1,p=-1;if(t&&s&&s.v7_partialHydration)for(let m=0;m<u.length;m++){let _=u[m];if((_.route.HydrateFallback||_.route.hydrateFallbackElement)&&(p=m),_.route.id){let{loaderData:x,errors:S}=t,T=_.route.loader&&x[_.route.id]===void 0&&(!S||S[_.route.id]===void 0);if(_.route.lazy||T){h=!0,p>=0?u=u.slice(0,p+1):u=[u[0]];break}}}return u.reduceRight((m,_,x)=>{let S,T=!1,w=null,v=null;t&&(S=f&&_.route.id?f[_.route.id]:void 0,w=_.route.errorElement||F0,h&&(p<0&&x===0?(j0("route-fallback"),T=!0,v=null):p===x&&(T=!0,v=_.route.hydrateFallbackElement||null)));let y=e.concat(u.slice(0,x+1)),P=()=>{let b;return S?b=w:T?b=v:_.route.Component?b=he.createElement(_.route.Component,null):_.route.element?b=_.route.element:b=m,he.createElement(k0,{match:_,routeContext:{outlet:m,matches:y,isDataRoute:t!=null},children:b})};return t&&(_.route.ErrorBoundary||_.route.errorElement||x===0)?he.createElement(O0,{location:t.location,revalidation:t.revalidation,component:w,error:S,children:P(),routeContext:{outlet:null,matches:y,isDataRoute:!0}}):P()},null)}var Cg=(function(r){return r.UseBlocker="useBlocker",r.UseRevalidator="useRevalidator",r.UseNavigateStable="useNavigate",r})(Cg||{}),Rg=(function(r){return r.UseBlocker="useBlocker",r.UseLoaderData="useLoaderData",r.UseActionData="useActionData",r.UseRouteError="useRouteError",r.UseNavigation="useNavigation",r.UseRouteLoaderData="useRouteLoaderData",r.UseMatches="useMatches",r.UseRevalidator="useRevalidator",r.UseNavigateStable="useNavigate",r.UseRouteId="useRouteId",r})(Rg||{});function z0(r){let e=he.useContext(pd);return e||Yt(!1),e}function H0(r){let e=he.useContext(P0);return e||Yt(!1),e}function V0(r){let e=he.useContext(Rr);return e||Yt(!1),e}function bg(r){let e=V0(),t=e.matches[e.matches.length-1];return t.route.id||Yt(!1),t.route.id}function G0(){var r;let e=he.useContext(Tg),t=H0(),s=bg();return e!==void 0?e:(r=t.errors)==null?void 0:r[s]}function W0(){let{router:r}=z0(Cg.UseNavigateStable),e=bg(Rg.UseNavigateStable),t=he.useRef(!1);return wg(()=>{t.current=!0}),he.useCallback(function(a,l){l===void 0&&(l={}),t.current&&(typeof a=="number"?r.navigate(a):r.navigate(a,ea({fromRouteId:e},l)))},[r,e])}const fm={};function j0(r,e,t){fm[r]||(fm[r]=!0)}function Kr(r){Yt(!1)}function X0(r){let{basename:e="/",children:t=null,location:s,navigationType:a=Mr.Pop,navigator:l,static:u=!1,future:f}=r;ia()&&Yt(!1);let h=e.replace(/^\/*/,"/"),p=he.useMemo(()=>({basename:h,navigator:l,static:u,future:ea({v7_relativeSplatPath:!1},f)}),[h,f,l,u]);typeof s=="string"&&(s=oo(s));let{pathname:m="/",search:_="",hash:x="",state:S=null,key:T="default"}=s,w=he.useMemo(()=>{let v=hd(m,h);return v==null?null:{location:{pathname:v,search:_,hash:x,state:S,key:T},navigationType:a}},[h,m,_,x,S,T,a]);return w==null?null:he.createElement(ss.Provider,{value:p},he.createElement(Ql.Provider,{children:t,value:w}))}function Y0(r){let{children:e,location:t}=r;return N0(Mf(e),t)}new Promise(()=>{});function Mf(r,e){e===void 0&&(e=[]);let t=[];return he.Children.forEach(r,(s,a)=>{if(!he.isValidElement(s))return;let l=[...e,a];if(s.type===he.Fragment){t.push.apply(t,Mf(s.props.children,l));return}s.type!==Kr&&Yt(!1),!s.props.index||!s.props.children||Yt(!1);let u={id:s.props.id||l.join("-"),caseSensitive:s.props.caseSensitive,element:s.props.element,Component:s.props.Component,index:s.props.index,path:s.props.path,loader:s.props.loader,action:s.props.action,errorElement:s.props.errorElement,ErrorBoundary:s.props.ErrorBoundary,hasErrorBoundary:s.props.ErrorBoundary!=null||s.props.errorElement!=null,shouldRevalidate:s.props.shouldRevalidate,handle:s.props.handle,lazy:s.props.lazy};s.props.children&&(u.children=Mf(s.props.children,l)),t.push(u)}),t}function Ef(){return Ef=Object.assign?Object.assign.bind():function(r){for(var e=1;e<arguments.length;e++){var t=arguments[e];for(var s in t)Object.prototype.hasOwnProperty.call(t,s)&&(r[s]=t[s])}return r},Ef.apply(this,arguments)}function q0(r,e){if(r==null)return{};var t={},s=Object.keys(r),a,l;for(l=0;l<s.length;l++)a=s[l],!(e.indexOf(a)>=0)&&(t[a]=r[a]);return t}function $0(r){return!!(r.metaKey||r.altKey||r.ctrlKey||r.shiftKey)}function K0(r,e){return r.button===0&&(!e||e==="_self")&&!$0(r)}function Tf(r){return r===void 0&&(r=""),new URLSearchParams(typeof r=="string"||Array.isArray(r)||r instanceof URLSearchParams?r:Object.keys(r).reduce((e,t)=>{let s=r[t];return e.concat(Array.isArray(s)?s.map(a=>[t,a]):[[t,s]])},[]))}function Z0(r,e){let t=Tf(r);return e&&e.forEach((s,a)=>{t.has(a)||e.getAll(a).forEach(l=>{t.append(a,l)})}),t}const Q0=["onClick","relative","reloadDocument","replace","state","target","to","preventScrollReset","unstable_viewTransition"],J0="6";try{window.__reactRouterVersion=J0}catch{}const ex="startTransition",dm=Jv[ex];function tx(r){let{basename:e,children:t,future:s,window:a}=r,l=he.useRef();l.current==null&&(l.current=s0({window:a,v5Compat:!0}));let u=l.current,[f,h]=he.useState({action:u.action,location:u.location}),{v7_startTransition:p}=s||{},m=he.useCallback(_=>{p&&dm?dm(()=>h(_)):h(_)},[h,p]);return he.useLayoutEffect(()=>u.listen(m),[u,m]),he.createElement(X0,{basename:e,children:t,location:f.location,navigationType:f.action,navigator:u,future:s})}const nx=typeof window<"u"&&typeof window.document<"u"&&typeof window.document.createElement<"u",ix=/^(?:[a-z][a-z0-9+.-]*:|\/\/)/i,yi=he.forwardRef(function(e,t){let{onClick:s,relative:a,reloadDocument:l,replace:u,state:f,target:h,to:p,preventScrollReset:m,unstable_viewTransition:_}=e,x=q0(e,Q0),{basename:S}=he.useContext(ss),T,w=!1;if(typeof p=="string"&&ix.test(p)&&(T=p,nx))try{let b=new URL(window.location.href),D=p.startsWith("//")?new URL(b.protocol+p):new URL(p),V=hd(D.pathname,S);D.origin===b.origin&&V!=null?p=V+D.search+D.hash:w=!0}catch{}let v=L0(p,{relative:a}),y=rx(p,{replace:u,state:f,target:h,preventScrollReset:m,relative:a,unstable_viewTransition:_});function P(b){s&&s(b),b.defaultPrevented||y(b)}return he.createElement("a",Ef({},x,{href:T||v,onClick:w||l?s:P,ref:t,target:h}))});var hm;(function(r){r.UseScrollRestoration="useScrollRestoration",r.UseSubmit="useSubmit",r.UseSubmitFetcher="useSubmitFetcher",r.UseFetcher="useFetcher",r.useViewTransitionState="useViewTransitionState"})(hm||(hm={}));var pm;(function(r){r.UseFetcher="useFetcher",r.UseFetchers="useFetchers",r.UseScrollRestoration="useScrollRestoration"})(pm||(pm={}));function rx(r,e){let{target:t,replace:s,state:a,preventScrollReset:l,relative:u,unstable_viewTransition:f}=e===void 0?{}:e,h=md(),p=ra(),m=Ag(r,{relative:u});return he.useCallback(_=>{if(K0(_,t)){_.preventDefault();let x=s!==void 0?s:Xl(p)===Xl(m);h(r,{replace:x,state:a,preventScrollReset:l,relative:u,unstable_viewTransition:f})}},[p,h,m,s,a,t,r,l,u,f])}function sx(r){let e=he.useRef(Tf(r)),t=he.useRef(!1),s=ra(),a=he.useMemo(()=>Z0(s.search,t.current?null:e.current),[s.search]),l=md(),u=he.useCallback((f,h)=>{const p=Tf(typeof f=="function"?f(a):f);t.current=!0,l("?"+p,h)},[l,a]);return[a,u]}const Pg="bgi_cart";function jo(){try{const r=localStorage.getItem(Pg);return r?JSON.parse(r):[]}catch{return[]}}function pl(r){localStorage.setItem(Pg,JSON.stringify(r)),window.dispatchEvent(new Event("bgi-cart-changed"))}function Ns(r){return[r.productSlug,r.configurationId].join("|")}function ec(){const[r,e]=he.useState(()=>jo());he.useEffect(()=>{const u=()=>e(jo());return window.addEventListener("bgi-cart-changed",u),window.addEventListener("storage",u),()=>{window.removeEventListener("bgi-cart-changed",u),window.removeEventListener("storage",u)}},[]);const t=he.useCallback(u=>{const f=jo(),h=f.findIndex(m=>Ns(m)===Ns(u));let p;h>=0?(p=[...f],p[h]={...p[h],quantity:p[h].quantity+u.quantity}):p=[...f,u],pl(p)},[]),s=he.useCallback(u=>{const f=jo().filter(h=>Ns(h)!==Ns(u));pl(f)},[]),a=he.useCallback((u,f)=>{const h=jo().map(p=>Ns(p)===Ns(u)?{...p,quantity:f}:p);pl(h.filter(p=>p.quantity>0))},[]),l=he.useCallback(()=>pl([]),[]);return{items:r,addItem:t,removeItem:s,setQuantity:a,clear:l}}const ox="https://api.unclaimedstreets.com/api/bgi";class Lg extends Error{status;constructor(e,t){super(t),this.status=e}}async function Bn(r,e){const t=await fetch(`${ox}${r}`,{...e,headers:{"Content-Type":"application/json",...e?.headers??{}}});if(!t.ok){let s=t.statusText;try{s=(await t.json()).error??s}catch{}throw new Lg(t.status,s)}if(t.status!==204)return t.json()}const Pt={listGames:()=>Bn("/games"),getGame:r=>Bn(`/games/${r}`),getGameSchema:r=>Bn(`/games/${r}/schema`),getGameCompatibility:r=>Bn(`/games/${r}/compatibility`),listSleeveProfiles:()=>Bn("/sleeve-profiles"),preview:(r,e,t)=>Bn("/configure/preview",{method:"POST",body:JSON.stringify({productSlug:r,params:e,sessionId:t})}),validate:(r,e,t)=>Bn("/configure/validate",{method:"POST",body:JSON.stringify({productSlug:r,params:e,sessionId:t})}),getConfiguration:r=>Bn(`/configurations/${r}`),cart:r=>Bn("/cart",{method:"POST",body:JSON.stringify({items:r})}),checkout:(r,e,t,s,a)=>Bn("/checkout",{method:"POST",body:JSON.stringify({items:r,customerEmail:e,successUrl:t,cancelUrl:s,sessionId:a})}),getOrder:r=>Bn(`/orders/${r}`),submitWaitlist:(r,e,t)=>Bn("/waitlist",{method:"POST",body:JSON.stringify({email:r,requestedGame:e,sessionId:t})}),recordEvent:(r,e={})=>Bn("/events",{method:"POST",body:JSON.stringify({eventType:r,...e})}).catch(()=>{})},mm="bgi_session_id";function ta(){let r=localStorage.getItem(mm);return r||(r=crypto.randomUUID(),localStorage.setItem(mm,r)),r}function ax(){const[r,e]=he.useState([]),[t,s]=he.useState(null);return he.useEffect(()=>{Pt.listGames().then(e).catch(a=>s(String(a)))},[]),k.jsxs("div",{className:"space-y-16",children:[k.jsx("section",{className:"relative -mt-10 overflow-hidden bg-bgi-hero px-6 pb-6 pt-16 sm:px-12",children:k.jsxs("div",{className:"relative max-w-xl",children:[k.jsx("span",{className:"pill border-2 border-bgi-ink bg-bgi-glow text-bgi-ink",children:"Printed to order · fit to your box"}),k.jsxs("h1",{className:"font-display mt-5 text-5xl font-bold leading-[1.05] text-bgi-ink sm:text-6xl",children:["Tray sets that",k.jsx("br",{}),k.jsxs("span",{className:"relative inline-block",children:["actually fit",k.jsx("svg",{"aria-hidden":"true",viewBox:"0 0 300 20",className:"absolute -bottom-2 left-0 h-4 w-full text-bgi-coral",children:k.jsx("path",{d:"M2 12 C 60 2, 120 18, 180 8 S 260 4, 298 10",fill:"none",stroke:"currentColor",strokeWidth:"6",strokeLinecap:"round"})})]})]}),k.jsx("p",{className:"mb-8 mt-6 max-w-md text-lg text-bgi-ink/70",children:"Sized to your sleeve class and your box — not a one-size insert. Pick your game below to start."})]})}),t&&k.jsx("p",{className:"text-red-600",children:t}),k.jsxs("section",{children:[k.jsx("h2",{className:"font-display mb-5 text-2xl font-bold text-bgi-ink",children:"Supported games"}),k.jsx("div",{className:"grid grid-cols-1 gap-5 sm:grid-cols-2 md:grid-cols-3",children:r.map(a=>k.jsx(lx,{game:a},a.slug))}),k.jsxs("p",{className:"mt-6 text-sm text-bgi-ink/60",children:["Don't see your game? ",k.jsx(cx,{})]})]})]})}function lx({game:r}){const e=r.productSlug?`/configure/${r.slug}`:`/games/${r.slug}`;return k.jsxs(yi,{to:e,onClick:()=>Pt.recordEvent("game_selected",{sessionId:ta(),gameSlug:r.slug}),className:"card card-hover group block p-5",children:[k.jsx("h3",{className:"font-display font-bold text-bgi-ink",children:r.name}),k.jsxs("p",{className:"mt-1 text-sm text-bgi-ink/60",children:[r.publisher,r.yearPublished?` · ${r.yearPublished}`:""]}),k.jsx("p",{className:"pill mt-3 border-2 border-bgi-ink/15 bg-white text-bgi-ink",children:"built to order"})]})}function cx(){const[r,e]=he.useState(""),[t,s]=he.useState(""),[a,l]=he.useState(!1);return a?k.jsx("span",{className:"text-bgi-teal",children:"Thanks — we'll let you know."}):k.jsxs("span",{className:"inline-flex flex-wrap items-center gap-2 align-middle",children:[k.jsx("input",{type:"text",placeholder:"Which game?",value:t,onChange:u=>s(u.target.value),className:"input-field inline w-40 py-1 text-sm"}),k.jsx("input",{type:"email",placeholder:"you@example.com",value:r,onChange:u=>e(u.target.value),className:"input-field inline w-48 py-1 text-sm"}),k.jsx("button",{className:"btn-secondary py-1 text-sm",onClick:()=>{!r||!t||Pt.submitWaitlist(r,t,ta()).then(()=>l(!0))},children:"Request it"})]})}function ux(){const{slug:r}=Jl(),[e,t]=he.useState(null);if(he.useEffect(()=>{r&&Pt.getGame(r).then(t)},[r]),!e)return k.jsx("p",{className:"text-bgi-ink/60",children:"Loading…"});const s=e.boxProfiles.some(a=>!a.verified)||e.sleeveProfiles.some(a=>!a.verified)||e.manifest.some(a=>!a.verified);return k.jsxs("div",{className:"card mx-auto max-w-lg space-y-4 p-6",children:[k.jsx("h1",{className:"font-display text-2xl font-bold text-bgi-ink",children:e.name}),k.jsxs("p",{className:"text-bgi-ink/70",children:[e.publisher,e.yearPublished?` · ${e.yearPublished}`:""]}),s&&k.jsxs("p",{className:"banner-caution",children:["Box and sleeve dimensions for this game are sourced from published/community references and pending physical verification."," ",k.jsx(yi,{to:`/games/${r}/compatibility`,className:"underline decoration-bgi-coral/50 underline-offset-2",children:"Read more"}),"."]}),e.productSlug?k.jsx(yi,{to:`/configure/${e.slug}`,className:"btn-primary w-full",children:"Configure your tray set"}):k.jsx("p",{className:"text-sm text-bgi-ink/60",children:"No configurable product yet for this game."}),k.jsxs("p",{className:"text-xs text-bgi-ink/50",children:["Unaffiliated with and unendorsed by ",e.publisher||"the publisher",". See"," ",k.jsx(yi,{to:`/games/${r}/compatibility`,className:"underline decoration-bgi-teal/40 underline-offset-2",children:"compatibility notes"}),"."]})]})}function fx({schema:r,values:e,onChange:t,sleeveProfiles:s,boxProfiles:a,errors:l,derived:u,derivedMax:f,derivedMin:h}){const p=Object.keys(r.properties);return k.jsx("div",{className:"space-y-6",children:p.map(m=>{const _=r.properties[m],x=_["x-label"]??m,S=_["x-helpText"],T=_["x-diagramAsset"],w=l?.[m],v=u?.[m];if(v!==void 0)return k.jsx(Gr,{label:x,helpText:S,diagram:T,error:w,children:k.jsxs("div",{className:"rounded-xl border border-bgi-teal/20 bg-bgi-foam px-3 py-2 text-bgi-ink/70",children:[v," ",k.jsx("span",{className:"text-xs",children:"(derived)"})]})},m);if(_["x-control"]==="sleeve-select")return k.jsx(Gr,{label:x,helpText:S,diagram:T,error:w,children:k.jsxs("select",{className:"input-field",value:e[m]??"",onChange:P=>t(m,P.target.value),children:[k.jsx("option",{value:"",disabled:!0,children:"Choose a sleeve class…"}),(s??[]).map(P=>k.jsxs("option",{value:P.id,children:[P.label,P.verified?"":" (unverified)"]},P.id))]})},m);if(_["x-control"]==="box-select")return k.jsx(Gr,{label:x,helpText:S,diagram:T,error:w,children:k.jsxs("select",{className:"input-field",value:e[m]??"",onChange:P=>t(m,P.target.value),children:[k.jsx("option",{value:"",disabled:!0,children:"Choose a box…"}),(a??[]).map(P=>k.jsxs("option",{value:P.id,children:[P.label," (",P.source,")",P.verified?"":" — unverified"]},P.id))]})},m);const y=dx(_.type);if(y==="boolean")return k.jsx(Gr,{label:x,helpText:S,diagram:T,error:w,children:k.jsx("input",{type:"checkbox",checked:!!e[m],onChange:P=>t(m,P.target.checked),className:"h-5 w-5"})},m);if(_.enum&&_.enum.length>0)return k.jsx(Gr,{label:x,helpText:S,diagram:T,error:w,children:k.jsx("select",{className:"input-field",value:String(e[m]??_.default??""),onChange:P=>t(m,hx(_.enum,P.target.value)),children:_.enum.map(P=>k.jsxs("option",{value:String(P),children:[px(P),_["x-unit"]?_["x-unit"]:""]},String(P)))})},m);if(y==="number"||y==="integer"){const P=f?.[m],b=h?.[m],D=P!==void 0?Math.min(_.maximum??P,P):_.maximum,V=b!==void 0?Math.max(_.minimum??b,b):_.minimum,O=P!==void 0&&D===P&&P<(_.maximum??1/0),U=b!==void 0&&V===b&&b>(_.minimum??-1/0);return k.jsxs(Gr,{label:x,helpText:S,diagram:T,error:w,children:[k.jsxs("div",{className:"flex items-center gap-2",children:[k.jsx("input",{type:"range",min:V,max:D,step:y==="integer"?1:.5,value:Number(e[m]??V??0),onChange:Y=>t(m,Number(Y.target.value)),className:"flex-1"}),k.jsxs("span",{className:"w-20 text-right text-sm tabular-nums",children:[Number(e[m]??V??0),_["x-unit"]??""]})]}),O||U?k.jsx("p",{className:"text-xs text-bgi-teal mt-1",children:O&&U?`Between ${b} and ${P} with your other settings.`:O?`Max ${P} with your other settings.`:`Min ${b} with your other settings.`}):(_.minimum!==void 0||_.maximum!==void 0)&&k.jsxs("p",{className:"text-xs text-bgi-ink/50 mt-1",children:["Range: ",_.minimum??"–"," to ",_.maximum??"–",_["x-unit"]??""]})]},m)}return k.jsx(Gr,{label:x,helpText:S,diagram:T,error:w,children:k.jsx("input",{type:"text",className:"w-full border border-bgi-teal/30 rounded px-3 py-2",value:String(e[m]??""),onChange:P=>t(m,P.target.value)})},m)})})}function Gr({label:r,helpText:e,diagram:t,error:s,children:a}){return k.jsxs("div",{children:[k.jsxs("div",{className:"mb-1 flex items-center justify-between",children:[k.jsx("label",{className:"text-sm font-medium text-bgi-ink",children:r}),t&&k.jsx("a",{href:t,target:"_blank",rel:"noreferrer",className:"text-xs font-medium text-bgi-teal underline decoration-bgi-teal/40 underline-offset-2 hover:text-bgi-coral",children:"where to measure"})]}),a,e&&k.jsx("p",{className:"text-xs text-bgi-ink/50 mt-1",children:e}),s&&k.jsxs("p",{className:"text-xs text-red-600 mt-1",children:[r,": ",s]})]})}function dx(r){return typeof r=="string"?r:r.find(e=>e&&e!=="null")??""}function hx(r,e){const t=r.find(s=>typeof s=="number"&&String(s)===e);return t!==void 0?t:e}function px(r){const e=String(r);return e.charAt(0).toUpperCase()+e.slice(1)}const gd="169",Ks={ROTATE:0,DOLLY:1,PAN:2},qs={ROTATE:0,PAN:1,DOLLY_PAN:2,DOLLY_ROTATE:3},mx=0,gm=1,gx=2,Dg=1,_x=2,Bi=3,Ar=0,Ln=1,zi=2,Tr=0,Zs=1,_m=2,vm=3,xm=4,vx=5,Qr=100,xx=101,yx=102,Sx=103,Mx=104,Ex=200,Tx=201,wx=202,Ax=203,wf=204,Af=205,Cx=206,Rx=207,bx=208,Px=209,Lx=210,Dx=211,Nx=212,Ix=213,Ux=214,Cf=0,Rf=1,bf=2,eo=3,Pf=4,Lf=5,Df=6,Nf=7,Ng=0,Fx=1,Ox=2,wr=0,kx=1,Bx=2,zx=3,Hx=4,Vx=5,Gx=6,Wx=7,Ig=300,to=301,no=302,If=303,Uf=304,tc=306,Ff=1e3,es=1001,Of=1002,Qn=1003,jx=1004,ml=1005,ui=1006,Vu=1007,ts=1008,Wi=1009,Ug=1010,Fg=1011,na=1012,_d=1013,ns=1014,Hi=1015,sa=1016,vd=1017,xd=1018,io=1020,Og=35902,kg=1021,Bg=1022,di=1023,zg=1024,Hg=1025,Qs=1026,ro=1027,Vg=1028,yd=1029,Gg=1030,Sd=1031,Md=1033,kl=33776,Bl=33777,zl=33778,Hl=33779,kf=35840,Bf=35841,zf=35842,Hf=35843,Vf=36196,Gf=37492,Wf=37496,jf=37808,Xf=37809,Yf=37810,qf=37811,$f=37812,Kf=37813,Zf=37814,Qf=37815,Jf=37816,ed=37817,td=37818,nd=37819,id=37820,rd=37821,Vl=36492,sd=36494,od=36495,Wg=36283,ad=36284,ld=36285,cd=36286,Xx=3200,Yx=3201,jg=0,qx=1,Sr="",ci="srgb",br="srgb-linear",Ed="display-p3",nc="display-p3-linear",Yl="linear",Ut="srgb",ql="rec709",$l="p3",Is=7680,ym=519,$x=512,Kx=513,Zx=514,Xg=515,Qx=516,Jx=517,ey=518,ty=519,Sm=35044,Mm="300 es",Vi=2e3,Kl=2001;class os{addEventListener(e,t){this._listeners===void 0&&(this._listeners={});const s=this._listeners;s[e]===void 0&&(s[e]=[]),s[e].indexOf(t)===-1&&s[e].push(t)}hasEventListener(e,t){if(this._listeners===void 0)return!1;const s=this._listeners;return s[e]!==void 0&&s[e].indexOf(t)!==-1}removeEventListener(e,t){if(this._listeners===void 0)return;const a=this._listeners[e];if(a!==void 0){const l=a.indexOf(t);l!==-1&&a.splice(l,1)}}dispatchEvent(e){if(this._listeners===void 0)return;const s=this._listeners[e.type];if(s!==void 0){e.target=this;const a=s.slice(0);for(let l=0,u=a.length;l<u;l++)a[l].call(this,e);e.target=null}}}const gn=["00","01","02","03","04","05","06","07","08","09","0a","0b","0c","0d","0e","0f","10","11","12","13","14","15","16","17","18","19","1a","1b","1c","1d","1e","1f","20","21","22","23","24","25","26","27","28","29","2a","2b","2c","2d","2e","2f","30","31","32","33","34","35","36","37","38","39","3a","3b","3c","3d","3e","3f","40","41","42","43","44","45","46","47","48","49","4a","4b","4c","4d","4e","4f","50","51","52","53","54","55","56","57","58","59","5a","5b","5c","5d","5e","5f","60","61","62","63","64","65","66","67","68","69","6a","6b","6c","6d","6e","6f","70","71","72","73","74","75","76","77","78","79","7a","7b","7c","7d","7e","7f","80","81","82","83","84","85","86","87","88","89","8a","8b","8c","8d","8e","8f","90","91","92","93","94","95","96","97","98","99","9a","9b","9c","9d","9e","9f","a0","a1","a2","a3","a4","a5","a6","a7","a8","a9","aa","ab","ac","ad","ae","af","b0","b1","b2","b3","b4","b5","b6","b7","b8","b9","ba","bb","bc","bd","be","bf","c0","c1","c2","c3","c4","c5","c6","c7","c8","c9","ca","cb","cc","cd","ce","cf","d0","d1","d2","d3","d4","d5","d6","d7","d8","d9","da","db","dc","dd","de","df","e0","e1","e2","e3","e4","e5","e6","e7","e8","e9","ea","eb","ec","ed","ee","ef","f0","f1","f2","f3","f4","f5","f6","f7","f8","f9","fa","fb","fc","fd","fe","ff"],Gl=Math.PI/180,ud=180/Math.PI;function oa(){const r=Math.random()*4294967295|0,e=Math.random()*4294967295|0,t=Math.random()*4294967295|0,s=Math.random()*4294967295|0;return(gn[r&255]+gn[r>>8&255]+gn[r>>16&255]+gn[r>>24&255]+"-"+gn[e&255]+gn[e>>8&255]+"-"+gn[e>>16&15|64]+gn[e>>24&255]+"-"+gn[t&63|128]+gn[t>>8&255]+"-"+gn[t>>16&255]+gn[t>>24&255]+gn[s&255]+gn[s>>8&255]+gn[s>>16&255]+gn[s>>24&255]).toLowerCase()}function Mn(r,e,t){return Math.max(e,Math.min(t,r))}function ny(r,e){return(r%e+e)%e}function Gu(r,e,t){return(1-t)*r+t*e}function Xo(r,e){switch(e.constructor){case Float32Array:return r;case Uint32Array:return r/4294967295;case Uint16Array:return r/65535;case Uint8Array:return r/255;case Int32Array:return Math.max(r/2147483647,-1);case Int16Array:return Math.max(r/32767,-1);case Int8Array:return Math.max(r/127,-1);default:throw new Error("Invalid component type.")}}function bn(r,e){switch(e.constructor){case Float32Array:return r;case Uint32Array:return Math.round(r*4294967295);case Uint16Array:return Math.round(r*65535);case Uint8Array:return Math.round(r*255);case Int32Array:return Math.round(r*2147483647);case Int16Array:return Math.round(r*32767);case Int8Array:return Math.round(r*127);default:throw new Error("Invalid component type.")}}const iy={DEG2RAD:Gl};class rt{constructor(e=0,t=0){rt.prototype.isVector2=!0,this.x=e,this.y=t}get width(){return this.x}set width(e){this.x=e}get height(){return this.y}set height(e){this.y=e}set(e,t){return this.x=e,this.y=t,this}setScalar(e){return this.x=e,this.y=e,this}setX(e){return this.x=e,this}setY(e){return this.y=e,this}setComponent(e,t){switch(e){case 0:this.x=t;break;case 1:this.y=t;break;default:throw new Error("index is out of range: "+e)}return this}getComponent(e){switch(e){case 0:return this.x;case 1:return this.y;default:throw new Error("index is out of range: "+e)}}clone(){return new this.constructor(this.x,this.y)}copy(e){return this.x=e.x,this.y=e.y,this}add(e){return this.x+=e.x,this.y+=e.y,this}addScalar(e){return this.x+=e,this.y+=e,this}addVectors(e,t){return this.x=e.x+t.x,this.y=e.y+t.y,this}addScaledVector(e,t){return this.x+=e.x*t,this.y+=e.y*t,this}sub(e){return this.x-=e.x,this.y-=e.y,this}subScalar(e){return this.x-=e,this.y-=e,this}subVectors(e,t){return this.x=e.x-t.x,this.y=e.y-t.y,this}multiply(e){return this.x*=e.x,this.y*=e.y,this}multiplyScalar(e){return this.x*=e,this.y*=e,this}divide(e){return this.x/=e.x,this.y/=e.y,this}divideScalar(e){return this.multiplyScalar(1/e)}applyMatrix3(e){const t=this.x,s=this.y,a=e.elements;return this.x=a[0]*t+a[3]*s+a[6],this.y=a[1]*t+a[4]*s+a[7],this}min(e){return this.x=Math.min(this.x,e.x),this.y=Math.min(this.y,e.y),this}max(e){return this.x=Math.max(this.x,e.x),this.y=Math.max(this.y,e.y),this}clamp(e,t){return this.x=Math.max(e.x,Math.min(t.x,this.x)),this.y=Math.max(e.y,Math.min(t.y,this.y)),this}clampScalar(e,t){return this.x=Math.max(e,Math.min(t,this.x)),this.y=Math.max(e,Math.min(t,this.y)),this}clampLength(e,t){const s=this.length();return this.divideScalar(s||1).multiplyScalar(Math.max(e,Math.min(t,s)))}floor(){return this.x=Math.floor(this.x),this.y=Math.floor(this.y),this}ceil(){return this.x=Math.ceil(this.x),this.y=Math.ceil(this.y),this}round(){return this.x=Math.round(this.x),this.y=Math.round(this.y),this}roundToZero(){return this.x=Math.trunc(this.x),this.y=Math.trunc(this.y),this}negate(){return this.x=-this.x,this.y=-this.y,this}dot(e){return this.x*e.x+this.y*e.y}cross(e){return this.x*e.y-this.y*e.x}lengthSq(){return this.x*this.x+this.y*this.y}length(){return Math.sqrt(this.x*this.x+this.y*this.y)}manhattanLength(){return Math.abs(this.x)+Math.abs(this.y)}normalize(){return this.divideScalar(this.length()||1)}angle(){return Math.atan2(-this.y,-this.x)+Math.PI}angleTo(e){const t=Math.sqrt(this.lengthSq()*e.lengthSq());if(t===0)return Math.PI/2;const s=this.dot(e)/t;return Math.acos(Mn(s,-1,1))}distanceTo(e){return Math.sqrt(this.distanceToSquared(e))}distanceToSquared(e){const t=this.x-e.x,s=this.y-e.y;return t*t+s*s}manhattanDistanceTo(e){return Math.abs(this.x-e.x)+Math.abs(this.y-e.y)}setLength(e){return this.normalize().multiplyScalar(e)}lerp(e,t){return this.x+=(e.x-this.x)*t,this.y+=(e.y-this.y)*t,this}lerpVectors(e,t,s){return this.x=e.x+(t.x-e.x)*s,this.y=e.y+(t.y-e.y)*s,this}equals(e){return e.x===this.x&&e.y===this.y}fromArray(e,t=0){return this.x=e[t],this.y=e[t+1],this}toArray(e=[],t=0){return e[t]=this.x,e[t+1]=this.y,e}fromBufferAttribute(e,t){return this.x=e.getX(t),this.y=e.getY(t),this}rotateAround(e,t){const s=Math.cos(t),a=Math.sin(t),l=this.x-e.x,u=this.y-e.y;return this.x=l*s-u*a+e.x,this.y=l*a+u*s+e.y,this}random(){return this.x=Math.random(),this.y=Math.random(),this}*[Symbol.iterator](){yield this.x,yield this.y}}class ot{constructor(e,t,s,a,l,u,f,h,p){ot.prototype.isMatrix3=!0,this.elements=[1,0,0,0,1,0,0,0,1],e!==void 0&&this.set(e,t,s,a,l,u,f,h,p)}set(e,t,s,a,l,u,f,h,p){const m=this.elements;return m[0]=e,m[1]=a,m[2]=f,m[3]=t,m[4]=l,m[5]=h,m[6]=s,m[7]=u,m[8]=p,this}identity(){return this.set(1,0,0,0,1,0,0,0,1),this}copy(e){const t=this.elements,s=e.elements;return t[0]=s[0],t[1]=s[1],t[2]=s[2],t[3]=s[3],t[4]=s[4],t[5]=s[5],t[6]=s[6],t[7]=s[7],t[8]=s[8],this}extractBasis(e,t,s){return e.setFromMatrix3Column(this,0),t.setFromMatrix3Column(this,1),s.setFromMatrix3Column(this,2),this}setFromMatrix4(e){const t=e.elements;return this.set(t[0],t[4],t[8],t[1],t[5],t[9],t[2],t[6],t[10]),this}multiply(e){return this.multiplyMatrices(this,e)}premultiply(e){return this.multiplyMatrices(e,this)}multiplyMatrices(e,t){const s=e.elements,a=t.elements,l=this.elements,u=s[0],f=s[3],h=s[6],p=s[1],m=s[4],_=s[7],x=s[2],S=s[5],T=s[8],w=a[0],v=a[3],y=a[6],P=a[1],b=a[4],D=a[7],V=a[2],O=a[5],U=a[8];return l[0]=u*w+f*P+h*V,l[3]=u*v+f*b+h*O,l[6]=u*y+f*D+h*U,l[1]=p*w+m*P+_*V,l[4]=p*v+m*b+_*O,l[7]=p*y+m*D+_*U,l[2]=x*w+S*P+T*V,l[5]=x*v+S*b+T*O,l[8]=x*y+S*D+T*U,this}multiplyScalar(e){const t=this.elements;return t[0]*=e,t[3]*=e,t[6]*=e,t[1]*=e,t[4]*=e,t[7]*=e,t[2]*=e,t[5]*=e,t[8]*=e,this}determinant(){const e=this.elements,t=e[0],s=e[1],a=e[2],l=e[3],u=e[4],f=e[5],h=e[6],p=e[7],m=e[8];return t*u*m-t*f*p-s*l*m+s*f*h+a*l*p-a*u*h}invert(){const e=this.elements,t=e[0],s=e[1],a=e[2],l=e[3],u=e[4],f=e[5],h=e[6],p=e[7],m=e[8],_=m*u-f*p,x=f*h-m*l,S=p*l-u*h,T=t*_+s*x+a*S;if(T===0)return this.set(0,0,0,0,0,0,0,0,0);const w=1/T;return e[0]=_*w,e[1]=(a*p-m*s)*w,e[2]=(f*s-a*u)*w,e[3]=x*w,e[4]=(m*t-a*h)*w,e[5]=(a*l-f*t)*w,e[6]=S*w,e[7]=(s*h-p*t)*w,e[8]=(u*t-s*l)*w,this}transpose(){let e;const t=this.elements;return e=t[1],t[1]=t[3],t[3]=e,e=t[2],t[2]=t[6],t[6]=e,e=t[5],t[5]=t[7],t[7]=e,this}getNormalMatrix(e){return this.setFromMatrix4(e).invert().transpose()}transposeIntoArray(e){const t=this.elements;return e[0]=t[0],e[1]=t[3],e[2]=t[6],e[3]=t[1],e[4]=t[4],e[5]=t[7],e[6]=t[2],e[7]=t[5],e[8]=t[8],this}setUvTransform(e,t,s,a,l,u,f){const h=Math.cos(l),p=Math.sin(l);return this.set(s*h,s*p,-s*(h*u+p*f)+u+e,-a*p,a*h,-a*(-p*u+h*f)+f+t,0,0,1),this}scale(e,t){return this.premultiply(Wu.makeScale(e,t)),this}rotate(e){return this.premultiply(Wu.makeRotation(-e)),this}translate(e,t){return this.premultiply(Wu.makeTranslation(e,t)),this}makeTranslation(e,t){return e.isVector2?this.set(1,0,e.x,0,1,e.y,0,0,1):this.set(1,0,e,0,1,t,0,0,1),this}makeRotation(e){const t=Math.cos(e),s=Math.sin(e);return this.set(t,-s,0,s,t,0,0,0,1),this}makeScale(e,t){return this.set(e,0,0,0,t,0,0,0,1),this}equals(e){const t=this.elements,s=e.elements;for(let a=0;a<9;a++)if(t[a]!==s[a])return!1;return!0}fromArray(e,t=0){for(let s=0;s<9;s++)this.elements[s]=e[s+t];return this}toArray(e=[],t=0){const s=this.elements;return e[t]=s[0],e[t+1]=s[1],e[t+2]=s[2],e[t+3]=s[3],e[t+4]=s[4],e[t+5]=s[5],e[t+6]=s[6],e[t+7]=s[7],e[t+8]=s[8],e}clone(){return new this.constructor().fromArray(this.elements)}}const Wu=new ot;function Yg(r){for(let e=r.length-1;e>=0;--e)if(r[e]>=65535)return!0;return!1}function Zl(r){return document.createElementNS("http://www.w3.org/1999/xhtml",r)}function ry(){const r=Zl("canvas");return r.style.display="block",r}const Em={};function Wl(r){r in Em||(Em[r]=!0,console.warn(r))}function sy(r,e,t){return new Promise(function(s,a){function l(){switch(r.clientWaitSync(e,r.SYNC_FLUSH_COMMANDS_BIT,0)){case r.WAIT_FAILED:a();break;case r.TIMEOUT_EXPIRED:setTimeout(l,t);break;default:s()}}setTimeout(l,t)})}function oy(r){const e=r.elements;e[2]=.5*e[2]+.5*e[3],e[6]=.5*e[6]+.5*e[7],e[10]=.5*e[10]+.5*e[11],e[14]=.5*e[14]+.5*e[15]}function ay(r){const e=r.elements;e[11]===-1?(e[10]=-e[10]-1,e[14]=-e[14]):(e[10]=-e[10],e[14]=-e[14]+1)}const Tm=new ot().set(.8224621,.177538,0,.0331941,.9668058,0,.0170827,.0723974,.9105199),wm=new ot().set(1.2249401,-.2249404,0,-.0420569,1.0420571,0,-.0196376,-.0786361,1.0982735),Yo={[br]:{transfer:Yl,primaries:ql,luminanceCoefficients:[.2126,.7152,.0722],toReference:r=>r,fromReference:r=>r},[ci]:{transfer:Ut,primaries:ql,luminanceCoefficients:[.2126,.7152,.0722],toReference:r=>r.convertSRGBToLinear(),fromReference:r=>r.convertLinearToSRGB()},[nc]:{transfer:Yl,primaries:$l,luminanceCoefficients:[.2289,.6917,.0793],toReference:r=>r.applyMatrix3(wm),fromReference:r=>r.applyMatrix3(Tm)},[Ed]:{transfer:Ut,primaries:$l,luminanceCoefficients:[.2289,.6917,.0793],toReference:r=>r.convertSRGBToLinear().applyMatrix3(wm),fromReference:r=>r.applyMatrix3(Tm).convertLinearToSRGB()}},ly=new Set([br,nc]),Tt={enabled:!0,_workingColorSpace:br,get workingColorSpace(){return this._workingColorSpace},set workingColorSpace(r){if(!ly.has(r))throw new Error(`Unsupported working color space, "${r}".`);this._workingColorSpace=r},convert:function(r,e,t){if(this.enabled===!1||e===t||!e||!t)return r;const s=Yo[e].toReference,a=Yo[t].fromReference;return a(s(r))},fromWorkingColorSpace:function(r,e){return this.convert(r,this._workingColorSpace,e)},toWorkingColorSpace:function(r,e){return this.convert(r,e,this._workingColorSpace)},getPrimaries:function(r){return Yo[r].primaries},getTransfer:function(r){return r===Sr?Yl:Yo[r].transfer},getLuminanceCoefficients:function(r,e=this._workingColorSpace){return r.fromArray(Yo[e].luminanceCoefficients)}};function Js(r){return r<.04045?r*.0773993808:Math.pow(r*.9478672986+.0521327014,2.4)}function ju(r){return r<.0031308?r*12.92:1.055*Math.pow(r,.41666)-.055}let Us;class cy{static getDataURL(e){if(/^data:/i.test(e.src)||typeof HTMLCanvasElement>"u")return e.src;let t;if(e instanceof HTMLCanvasElement)t=e;else{Us===void 0&&(Us=Zl("canvas")),Us.width=e.width,Us.height=e.height;const s=Us.getContext("2d");e instanceof ImageData?s.putImageData(e,0,0):s.drawImage(e,0,0,e.width,e.height),t=Us}return t.width>2048||t.height>2048?(console.warn("THREE.ImageUtils.getDataURL: Image converted to jpg for performance reasons",e),t.toDataURL("image/jpeg",.6)):t.toDataURL("image/png")}static sRGBToLinear(e){if(typeof HTMLImageElement<"u"&&e instanceof HTMLImageElement||typeof HTMLCanvasElement<"u"&&e instanceof HTMLCanvasElement||typeof ImageBitmap<"u"&&e instanceof ImageBitmap){const t=Zl("canvas");t.width=e.width,t.height=e.height;const s=t.getContext("2d");s.drawImage(e,0,0,e.width,e.height);const a=s.getImageData(0,0,e.width,e.height),l=a.data;for(let u=0;u<l.length;u++)l[u]=Js(l[u]/255)*255;return s.putImageData(a,0,0),t}else if(e.data){const t=e.data.slice(0);for(let s=0;s<t.length;s++)t instanceof Uint8Array||t instanceof Uint8ClampedArray?t[s]=Math.floor(Js(t[s]/255)*255):t[s]=Js(t[s]);return{data:t,width:e.width,height:e.height}}else return console.warn("THREE.ImageUtils.sRGBToLinear(): Unsupported image type. No color space conversion applied."),e}}let uy=0;class qg{constructor(e=null){this.isSource=!0,Object.defineProperty(this,"id",{value:uy++}),this.uuid=oa(),this.data=e,this.dataReady=!0,this.version=0}set needsUpdate(e){e===!0&&this.version++}toJSON(e){const t=e===void 0||typeof e=="string";if(!t&&e.images[this.uuid]!==void 0)return e.images[this.uuid];const s={uuid:this.uuid,url:""},a=this.data;if(a!==null){let l;if(Array.isArray(a)){l=[];for(let u=0,f=a.length;u<f;u++)a[u].isDataTexture?l.push(Xu(a[u].image)):l.push(Xu(a[u]))}else l=Xu(a);s.url=l}return t||(e.images[this.uuid]=s),s}}function Xu(r){return typeof HTMLImageElement<"u"&&r instanceof HTMLImageElement||typeof HTMLCanvasElement<"u"&&r instanceof HTMLCanvasElement||typeof ImageBitmap<"u"&&r instanceof ImageBitmap?cy.getDataURL(r):r.data?{data:Array.from(r.data),width:r.width,height:r.height,type:r.data.constructor.name}:(console.warn("THREE.Texture: Unable to serialize Texture."),{})}let fy=0;class Dn extends os{constructor(e=Dn.DEFAULT_IMAGE,t=Dn.DEFAULT_MAPPING,s=es,a=es,l=ui,u=ts,f=di,h=Wi,p=Dn.DEFAULT_ANISOTROPY,m=Sr){super(),this.isTexture=!0,Object.defineProperty(this,"id",{value:fy++}),this.uuid=oa(),this.name="",this.source=new qg(e),this.mipmaps=[],this.mapping=t,this.channel=0,this.wrapS=s,this.wrapT=a,this.magFilter=l,this.minFilter=u,this.anisotropy=p,this.format=f,this.internalFormat=null,this.type=h,this.offset=new rt(0,0),this.repeat=new rt(1,1),this.center=new rt(0,0),this.rotation=0,this.matrixAutoUpdate=!0,this.matrix=new ot,this.generateMipmaps=!0,this.premultiplyAlpha=!1,this.flipY=!0,this.unpackAlignment=4,this.colorSpace=m,this.userData={},this.version=0,this.onUpdate=null,this.isRenderTargetTexture=!1,this.pmremVersion=0}get image(){return this.source.data}set image(e=null){this.source.data=e}updateMatrix(){this.matrix.setUvTransform(this.offset.x,this.offset.y,this.repeat.x,this.repeat.y,this.rotation,this.center.x,this.center.y)}clone(){return new this.constructor().copy(this)}copy(e){return this.name=e.name,this.source=e.source,this.mipmaps=e.mipmaps.slice(0),this.mapping=e.mapping,this.channel=e.channel,this.wrapS=e.wrapS,this.wrapT=e.wrapT,this.magFilter=e.magFilter,this.minFilter=e.minFilter,this.anisotropy=e.anisotropy,this.format=e.format,this.internalFormat=e.internalFormat,this.type=e.type,this.offset.copy(e.offset),this.repeat.copy(e.repeat),this.center.copy(e.center),this.rotation=e.rotation,this.matrixAutoUpdate=e.matrixAutoUpdate,this.matrix.copy(e.matrix),this.generateMipmaps=e.generateMipmaps,this.premultiplyAlpha=e.premultiplyAlpha,this.flipY=e.flipY,this.unpackAlignment=e.unpackAlignment,this.colorSpace=e.colorSpace,this.userData=JSON.parse(JSON.stringify(e.userData)),this.needsUpdate=!0,this}toJSON(e){const t=e===void 0||typeof e=="string";if(!t&&e.textures[this.uuid]!==void 0)return e.textures[this.uuid];const s={metadata:{version:4.6,type:"Texture",generator:"Texture.toJSON"},uuid:this.uuid,name:this.name,image:this.source.toJSON(e).uuid,mapping:this.mapping,channel:this.channel,repeat:[this.repeat.x,this.repeat.y],offset:[this.offset.x,this.offset.y],center:[this.center.x,this.center.y],rotation:this.rotation,wrap:[this.wrapS,this.wrapT],format:this.format,internalFormat:this.internalFormat,type:this.type,colorSpace:this.colorSpace,minFilter:this.minFilter,magFilter:this.magFilter,anisotropy:this.anisotropy,flipY:this.flipY,generateMipmaps:this.generateMipmaps,premultiplyAlpha:this.premultiplyAlpha,unpackAlignment:this.unpackAlignment};return Object.keys(this.userData).length>0&&(s.userData=this.userData),t||(e.textures[this.uuid]=s),s}dispose(){this.dispatchEvent({type:"dispose"})}transformUv(e){if(this.mapping!==Ig)return e;if(e.applyMatrix3(this.matrix),e.x<0||e.x>1)switch(this.wrapS){case Ff:e.x=e.x-Math.floor(e.x);break;case es:e.x=e.x<0?0:1;break;case Of:Math.abs(Math.floor(e.x)%2)===1?e.x=Math.ceil(e.x)-e.x:e.x=e.x-Math.floor(e.x);break}if(e.y<0||e.y>1)switch(this.wrapT){case Ff:e.y=e.y-Math.floor(e.y);break;case es:e.y=e.y<0?0:1;break;case Of:Math.abs(Math.floor(e.y)%2)===1?e.y=Math.ceil(e.y)-e.y:e.y=e.y-Math.floor(e.y);break}return this.flipY&&(e.y=1-e.y),e}set needsUpdate(e){e===!0&&(this.version++,this.source.needsUpdate=!0)}set needsPMREMUpdate(e){e===!0&&this.pmremVersion++}}Dn.DEFAULT_IMAGE=null;Dn.DEFAULT_MAPPING=Ig;Dn.DEFAULT_ANISOTROPY=1;class Vt{constructor(e=0,t=0,s=0,a=1){Vt.prototype.isVector4=!0,this.x=e,this.y=t,this.z=s,this.w=a}get width(){return this.z}set width(e){this.z=e}get height(){return this.w}set height(e){this.w=e}set(e,t,s,a){return this.x=e,this.y=t,this.z=s,this.w=a,this}setScalar(e){return this.x=e,this.y=e,this.z=e,this.w=e,this}setX(e){return this.x=e,this}setY(e){return this.y=e,this}setZ(e){return this.z=e,this}setW(e){return this.w=e,this}setComponent(e,t){switch(e){case 0:this.x=t;break;case 1:this.y=t;break;case 2:this.z=t;break;case 3:this.w=t;break;default:throw new Error("index is out of range: "+e)}return this}getComponent(e){switch(e){case 0:return this.x;case 1:return this.y;case 2:return this.z;case 3:return this.w;default:throw new Error("index is out of range: "+e)}}clone(){return new this.constructor(this.x,this.y,this.z,this.w)}copy(e){return this.x=e.x,this.y=e.y,this.z=e.z,this.w=e.w!==void 0?e.w:1,this}add(e){return this.x+=e.x,this.y+=e.y,this.z+=e.z,this.w+=e.w,this}addScalar(e){return this.x+=e,this.y+=e,this.z+=e,this.w+=e,this}addVectors(e,t){return this.x=e.x+t.x,this.y=e.y+t.y,this.z=e.z+t.z,this.w=e.w+t.w,this}addScaledVector(e,t){return this.x+=e.x*t,this.y+=e.y*t,this.z+=e.z*t,this.w+=e.w*t,this}sub(e){return this.x-=e.x,this.y-=e.y,this.z-=e.z,this.w-=e.w,this}subScalar(e){return this.x-=e,this.y-=e,this.z-=e,this.w-=e,this}subVectors(e,t){return this.x=e.x-t.x,this.y=e.y-t.y,this.z=e.z-t.z,this.w=e.w-t.w,this}multiply(e){return this.x*=e.x,this.y*=e.y,this.z*=e.z,this.w*=e.w,this}multiplyScalar(e){return this.x*=e,this.y*=e,this.z*=e,this.w*=e,this}applyMatrix4(e){const t=this.x,s=this.y,a=this.z,l=this.w,u=e.elements;return this.x=u[0]*t+u[4]*s+u[8]*a+u[12]*l,this.y=u[1]*t+u[5]*s+u[9]*a+u[13]*l,this.z=u[2]*t+u[6]*s+u[10]*a+u[14]*l,this.w=u[3]*t+u[7]*s+u[11]*a+u[15]*l,this}divideScalar(e){return this.multiplyScalar(1/e)}setAxisAngleFromQuaternion(e){this.w=2*Math.acos(e.w);const t=Math.sqrt(1-e.w*e.w);return t<1e-4?(this.x=1,this.y=0,this.z=0):(this.x=e.x/t,this.y=e.y/t,this.z=e.z/t),this}setAxisAngleFromRotationMatrix(e){let t,s,a,l;const h=e.elements,p=h[0],m=h[4],_=h[8],x=h[1],S=h[5],T=h[9],w=h[2],v=h[6],y=h[10];if(Math.abs(m-x)<.01&&Math.abs(_-w)<.01&&Math.abs(T-v)<.01){if(Math.abs(m+x)<.1&&Math.abs(_+w)<.1&&Math.abs(T+v)<.1&&Math.abs(p+S+y-3)<.1)return this.set(1,0,0,0),this;t=Math.PI;const b=(p+1)/2,D=(S+1)/2,V=(y+1)/2,O=(m+x)/4,U=(_+w)/4,Y=(T+v)/4;return b>D&&b>V?b<.01?(s=0,a=.707106781,l=.707106781):(s=Math.sqrt(b),a=O/s,l=U/s):D>V?D<.01?(s=.707106781,a=0,l=.707106781):(a=Math.sqrt(D),s=O/a,l=Y/a):V<.01?(s=.707106781,a=.707106781,l=0):(l=Math.sqrt(V),s=U/l,a=Y/l),this.set(s,a,l,t),this}let P=Math.sqrt((v-T)*(v-T)+(_-w)*(_-w)+(x-m)*(x-m));return Math.abs(P)<.001&&(P=1),this.x=(v-T)/P,this.y=(_-w)/P,this.z=(x-m)/P,this.w=Math.acos((p+S+y-1)/2),this}setFromMatrixPosition(e){const t=e.elements;return this.x=t[12],this.y=t[13],this.z=t[14],this.w=t[15],this}min(e){return this.x=Math.min(this.x,e.x),this.y=Math.min(this.y,e.y),this.z=Math.min(this.z,e.z),this.w=Math.min(this.w,e.w),this}max(e){return this.x=Math.max(this.x,e.x),this.y=Math.max(this.y,e.y),this.z=Math.max(this.z,e.z),this.w=Math.max(this.w,e.w),this}clamp(e,t){return this.x=Math.max(e.x,Math.min(t.x,this.x)),this.y=Math.max(e.y,Math.min(t.y,this.y)),this.z=Math.max(e.z,Math.min(t.z,this.z)),this.w=Math.max(e.w,Math.min(t.w,this.w)),this}clampScalar(e,t){return this.x=Math.max(e,Math.min(t,this.x)),this.y=Math.max(e,Math.min(t,this.y)),this.z=Math.max(e,Math.min(t,this.z)),this.w=Math.max(e,Math.min(t,this.w)),this}clampLength(e,t){const s=this.length();return this.divideScalar(s||1).multiplyScalar(Math.max(e,Math.min(t,s)))}floor(){return this.x=Math.floor(this.x),this.y=Math.floor(this.y),this.z=Math.floor(this.z),this.w=Math.floor(this.w),this}ceil(){return this.x=Math.ceil(this.x),this.y=Math.ceil(this.y),this.z=Math.ceil(this.z),this.w=Math.ceil(this.w),this}round(){return this.x=Math.round(this.x),this.y=Math.round(this.y),this.z=Math.round(this.z),this.w=Math.round(this.w),this}roundToZero(){return this.x=Math.trunc(this.x),this.y=Math.trunc(this.y),this.z=Math.trunc(this.z),this.w=Math.trunc(this.w),this}negate(){return this.x=-this.x,this.y=-this.y,this.z=-this.z,this.w=-this.w,this}dot(e){return this.x*e.x+this.y*e.y+this.z*e.z+this.w*e.w}lengthSq(){return this.x*this.x+this.y*this.y+this.z*this.z+this.w*this.w}length(){return Math.sqrt(this.x*this.x+this.y*this.y+this.z*this.z+this.w*this.w)}manhattanLength(){return Math.abs(this.x)+Math.abs(this.y)+Math.abs(this.z)+Math.abs(this.w)}normalize(){return this.divideScalar(this.length()||1)}setLength(e){return this.normalize().multiplyScalar(e)}lerp(e,t){return this.x+=(e.x-this.x)*t,this.y+=(e.y-this.y)*t,this.z+=(e.z-this.z)*t,this.w+=(e.w-this.w)*t,this}lerpVectors(e,t,s){return this.x=e.x+(t.x-e.x)*s,this.y=e.y+(t.y-e.y)*s,this.z=e.z+(t.z-e.z)*s,this.w=e.w+(t.w-e.w)*s,this}equals(e){return e.x===this.x&&e.y===this.y&&e.z===this.z&&e.w===this.w}fromArray(e,t=0){return this.x=e[t],this.y=e[t+1],this.z=e[t+2],this.w=e[t+3],this}toArray(e=[],t=0){return e[t]=this.x,e[t+1]=this.y,e[t+2]=this.z,e[t+3]=this.w,e}fromBufferAttribute(e,t){return this.x=e.getX(t),this.y=e.getY(t),this.z=e.getZ(t),this.w=e.getW(t),this}random(){return this.x=Math.random(),this.y=Math.random(),this.z=Math.random(),this.w=Math.random(),this}*[Symbol.iterator](){yield this.x,yield this.y,yield this.z,yield this.w}}class dy extends os{constructor(e=1,t=1,s={}){super(),this.isRenderTarget=!0,this.width=e,this.height=t,this.depth=1,this.scissor=new Vt(0,0,e,t),this.scissorTest=!1,this.viewport=new Vt(0,0,e,t);const a={width:e,height:t,depth:1};s=Object.assign({generateMipmaps:!1,internalFormat:null,minFilter:ui,depthBuffer:!0,stencilBuffer:!1,resolveDepthBuffer:!0,resolveStencilBuffer:!0,depthTexture:null,samples:0,count:1},s);const l=new Dn(a,s.mapping,s.wrapS,s.wrapT,s.magFilter,s.minFilter,s.format,s.type,s.anisotropy,s.colorSpace);l.flipY=!1,l.generateMipmaps=s.generateMipmaps,l.internalFormat=s.internalFormat,this.textures=[];const u=s.count;for(let f=0;f<u;f++)this.textures[f]=l.clone(),this.textures[f].isRenderTargetTexture=!0;this.depthBuffer=s.depthBuffer,this.stencilBuffer=s.stencilBuffer,this.resolveDepthBuffer=s.resolveDepthBuffer,this.resolveStencilBuffer=s.resolveStencilBuffer,this.depthTexture=s.depthTexture,this.samples=s.samples}get texture(){return this.textures[0]}set texture(e){this.textures[0]=e}setSize(e,t,s=1){if(this.width!==e||this.height!==t||this.depth!==s){this.width=e,this.height=t,this.depth=s;for(let a=0,l=this.textures.length;a<l;a++)this.textures[a].image.width=e,this.textures[a].image.height=t,this.textures[a].image.depth=s;this.dispose()}this.viewport.set(0,0,e,t),this.scissor.set(0,0,e,t)}clone(){return new this.constructor().copy(this)}copy(e){this.width=e.width,this.height=e.height,this.depth=e.depth,this.scissor.copy(e.scissor),this.scissorTest=e.scissorTest,this.viewport.copy(e.viewport),this.textures.length=0;for(let s=0,a=e.textures.length;s<a;s++)this.textures[s]=e.textures[s].clone(),this.textures[s].isRenderTargetTexture=!0;const t=Object.assign({},e.texture.image);return this.texture.source=new qg(t),this.depthBuffer=e.depthBuffer,this.stencilBuffer=e.stencilBuffer,this.resolveDepthBuffer=e.resolveDepthBuffer,this.resolveStencilBuffer=e.resolveStencilBuffer,e.depthTexture!==null&&(this.depthTexture=e.depthTexture.clone()),this.samples=e.samples,this}dispose(){this.dispatchEvent({type:"dispose"})}}class is extends dy{constructor(e=1,t=1,s={}){super(e,t,s),this.isWebGLRenderTarget=!0}}class $g extends Dn{constructor(e=null,t=1,s=1,a=1){super(null),this.isDataArrayTexture=!0,this.image={data:e,width:t,height:s,depth:a},this.magFilter=Qn,this.minFilter=Qn,this.wrapR=es,this.generateMipmaps=!1,this.flipY=!1,this.unpackAlignment=1,this.layerUpdates=new Set}addLayerUpdate(e){this.layerUpdates.add(e)}clearLayerUpdates(){this.layerUpdates.clear()}}class hy extends Dn{constructor(e=null,t=1,s=1,a=1){super(null),this.isData3DTexture=!0,this.image={data:e,width:t,height:s,depth:a},this.magFilter=Qn,this.minFilter=Qn,this.wrapR=es,this.generateMipmaps=!1,this.flipY=!1,this.unpackAlignment=1}}class rs{constructor(e=0,t=0,s=0,a=1){this.isQuaternion=!0,this._x=e,this._y=t,this._z=s,this._w=a}static slerpFlat(e,t,s,a,l,u,f){let h=s[a+0],p=s[a+1],m=s[a+2],_=s[a+3];const x=l[u+0],S=l[u+1],T=l[u+2],w=l[u+3];if(f===0){e[t+0]=h,e[t+1]=p,e[t+2]=m,e[t+3]=_;return}if(f===1){e[t+0]=x,e[t+1]=S,e[t+2]=T,e[t+3]=w;return}if(_!==w||h!==x||p!==S||m!==T){let v=1-f;const y=h*x+p*S+m*T+_*w,P=y>=0?1:-1,b=1-y*y;if(b>Number.EPSILON){const V=Math.sqrt(b),O=Math.atan2(V,y*P);v=Math.sin(v*O)/V,f=Math.sin(f*O)/V}const D=f*P;if(h=h*v+x*D,p=p*v+S*D,m=m*v+T*D,_=_*v+w*D,v===1-f){const V=1/Math.sqrt(h*h+p*p+m*m+_*_);h*=V,p*=V,m*=V,_*=V}}e[t]=h,e[t+1]=p,e[t+2]=m,e[t+3]=_}static multiplyQuaternionsFlat(e,t,s,a,l,u){const f=s[a],h=s[a+1],p=s[a+2],m=s[a+3],_=l[u],x=l[u+1],S=l[u+2],T=l[u+3];return e[t]=f*T+m*_+h*S-p*x,e[t+1]=h*T+m*x+p*_-f*S,e[t+2]=p*T+m*S+f*x-h*_,e[t+3]=m*T-f*_-h*x-p*S,e}get x(){return this._x}set x(e){this._x=e,this._onChangeCallback()}get y(){return this._y}set y(e){this._y=e,this._onChangeCallback()}get z(){return this._z}set z(e){this._z=e,this._onChangeCallback()}get w(){return this._w}set w(e){this._w=e,this._onChangeCallback()}set(e,t,s,a){return this._x=e,this._y=t,this._z=s,this._w=a,this._onChangeCallback(),this}clone(){return new this.constructor(this._x,this._y,this._z,this._w)}copy(e){return this._x=e.x,this._y=e.y,this._z=e.z,this._w=e.w,this._onChangeCallback(),this}setFromEuler(e,t=!0){const s=e._x,a=e._y,l=e._z,u=e._order,f=Math.cos,h=Math.sin,p=f(s/2),m=f(a/2),_=f(l/2),x=h(s/2),S=h(a/2),T=h(l/2);switch(u){case"XYZ":this._x=x*m*_+p*S*T,this._y=p*S*_-x*m*T,this._z=p*m*T+x*S*_,this._w=p*m*_-x*S*T;break;case"YXZ":this._x=x*m*_+p*S*T,this._y=p*S*_-x*m*T,this._z=p*m*T-x*S*_,this._w=p*m*_+x*S*T;break;case"ZXY":this._x=x*m*_-p*S*T,this._y=p*S*_+x*m*T,this._z=p*m*T+x*S*_,this._w=p*m*_-x*S*T;break;case"ZYX":this._x=x*m*_-p*S*T,this._y=p*S*_+x*m*T,this._z=p*m*T-x*S*_,this._w=p*m*_+x*S*T;break;case"YZX":this._x=x*m*_+p*S*T,this._y=p*S*_+x*m*T,this._z=p*m*T-x*S*_,this._w=p*m*_-x*S*T;break;case"XZY":this._x=x*m*_-p*S*T,this._y=p*S*_-x*m*T,this._z=p*m*T+x*S*_,this._w=p*m*_+x*S*T;break;default:console.warn("THREE.Quaternion: .setFromEuler() encountered an unknown order: "+u)}return t===!0&&this._onChangeCallback(),this}setFromAxisAngle(e,t){const s=t/2,a=Math.sin(s);return this._x=e.x*a,this._y=e.y*a,this._z=e.z*a,this._w=Math.cos(s),this._onChangeCallback(),this}setFromRotationMatrix(e){const t=e.elements,s=t[0],a=t[4],l=t[8],u=t[1],f=t[5],h=t[9],p=t[2],m=t[6],_=t[10],x=s+f+_;if(x>0){const S=.5/Math.sqrt(x+1);this._w=.25/S,this._x=(m-h)*S,this._y=(l-p)*S,this._z=(u-a)*S}else if(s>f&&s>_){const S=2*Math.sqrt(1+s-f-_);this._w=(m-h)/S,this._x=.25*S,this._y=(a+u)/S,this._z=(l+p)/S}else if(f>_){const S=2*Math.sqrt(1+f-s-_);this._w=(l-p)/S,this._x=(a+u)/S,this._y=.25*S,this._z=(h+m)/S}else{const S=2*Math.sqrt(1+_-s-f);this._w=(u-a)/S,this._x=(l+p)/S,this._y=(h+m)/S,this._z=.25*S}return this._onChangeCallback(),this}setFromUnitVectors(e,t){let s=e.dot(t)+1;return s<Number.EPSILON?(s=0,Math.abs(e.x)>Math.abs(e.z)?(this._x=-e.y,this._y=e.x,this._z=0,this._w=s):(this._x=0,this._y=-e.z,this._z=e.y,this._w=s)):(this._x=e.y*t.z-e.z*t.y,this._y=e.z*t.x-e.x*t.z,this._z=e.x*t.y-e.y*t.x,this._w=s),this.normalize()}angleTo(e){return 2*Math.acos(Math.abs(Mn(this.dot(e),-1,1)))}rotateTowards(e,t){const s=this.angleTo(e);if(s===0)return this;const a=Math.min(1,t/s);return this.slerp(e,a),this}identity(){return this.set(0,0,0,1)}invert(){return this.conjugate()}conjugate(){return this._x*=-1,this._y*=-1,this._z*=-1,this._onChangeCallback(),this}dot(e){return this._x*e._x+this._y*e._y+this._z*e._z+this._w*e._w}lengthSq(){return this._x*this._x+this._y*this._y+this._z*this._z+this._w*this._w}length(){return Math.sqrt(this._x*this._x+this._y*this._y+this._z*this._z+this._w*this._w)}normalize(){let e=this.length();return e===0?(this._x=0,this._y=0,this._z=0,this._w=1):(e=1/e,this._x=this._x*e,this._y=this._y*e,this._z=this._z*e,this._w=this._w*e),this._onChangeCallback(),this}multiply(e){return this.multiplyQuaternions(this,e)}premultiply(e){return this.multiplyQuaternions(e,this)}multiplyQuaternions(e,t){const s=e._x,a=e._y,l=e._z,u=e._w,f=t._x,h=t._y,p=t._z,m=t._w;return this._x=s*m+u*f+a*p-l*h,this._y=a*m+u*h+l*f-s*p,this._z=l*m+u*p+s*h-a*f,this._w=u*m-s*f-a*h-l*p,this._onChangeCallback(),this}slerp(e,t){if(t===0)return this;if(t===1)return this.copy(e);const s=this._x,a=this._y,l=this._z,u=this._w;let f=u*e._w+s*e._x+a*e._y+l*e._z;if(f<0?(this._w=-e._w,this._x=-e._x,this._y=-e._y,this._z=-e._z,f=-f):this.copy(e),f>=1)return this._w=u,this._x=s,this._y=a,this._z=l,this;const h=1-f*f;if(h<=Number.EPSILON){const S=1-t;return this._w=S*u+t*this._w,this._x=S*s+t*this._x,this._y=S*a+t*this._y,this._z=S*l+t*this._z,this.normalize(),this}const p=Math.sqrt(h),m=Math.atan2(p,f),_=Math.sin((1-t)*m)/p,x=Math.sin(t*m)/p;return this._w=u*_+this._w*x,this._x=s*_+this._x*x,this._y=a*_+this._y*x,this._z=l*_+this._z*x,this._onChangeCallback(),this}slerpQuaternions(e,t,s){return this.copy(e).slerp(t,s)}random(){const e=2*Math.PI*Math.random(),t=2*Math.PI*Math.random(),s=Math.random(),a=Math.sqrt(1-s),l=Math.sqrt(s);return this.set(a*Math.sin(e),a*Math.cos(e),l*Math.sin(t),l*Math.cos(t))}equals(e){return e._x===this._x&&e._y===this._y&&e._z===this._z&&e._w===this._w}fromArray(e,t=0){return this._x=e[t],this._y=e[t+1],this._z=e[t+2],this._w=e[t+3],this._onChangeCallback(),this}toArray(e=[],t=0){return e[t]=this._x,e[t+1]=this._y,e[t+2]=this._z,e[t+3]=this._w,e}fromBufferAttribute(e,t){return this._x=e.getX(t),this._y=e.getY(t),this._z=e.getZ(t),this._w=e.getW(t),this._onChangeCallback(),this}toJSON(){return this.toArray()}_onChange(e){return this._onChangeCallback=e,this}_onChangeCallback(){}*[Symbol.iterator](){yield this._x,yield this._y,yield this._z,yield this._w}}class Q{constructor(e=0,t=0,s=0){Q.prototype.isVector3=!0,this.x=e,this.y=t,this.z=s}set(e,t,s){return s===void 0&&(s=this.z),this.x=e,this.y=t,this.z=s,this}setScalar(e){return this.x=e,this.y=e,this.z=e,this}setX(e){return this.x=e,this}setY(e){return this.y=e,this}setZ(e){return this.z=e,this}setComponent(e,t){switch(e){case 0:this.x=t;break;case 1:this.y=t;break;case 2:this.z=t;break;default:throw new Error("index is out of range: "+e)}return this}getComponent(e){switch(e){case 0:return this.x;case 1:return this.y;case 2:return this.z;default:throw new Error("index is out of range: "+e)}}clone(){return new this.constructor(this.x,this.y,this.z)}copy(e){return this.x=e.x,this.y=e.y,this.z=e.z,this}add(e){return this.x+=e.x,this.y+=e.y,this.z+=e.z,this}addScalar(e){return this.x+=e,this.y+=e,this.z+=e,this}addVectors(e,t){return this.x=e.x+t.x,this.y=e.y+t.y,this.z=e.z+t.z,this}addScaledVector(e,t){return this.x+=e.x*t,this.y+=e.y*t,this.z+=e.z*t,this}sub(e){return this.x-=e.x,this.y-=e.y,this.z-=e.z,this}subScalar(e){return this.x-=e,this.y-=e,this.z-=e,this}subVectors(e,t){return this.x=e.x-t.x,this.y=e.y-t.y,this.z=e.z-t.z,this}multiply(e){return this.x*=e.x,this.y*=e.y,this.z*=e.z,this}multiplyScalar(e){return this.x*=e,this.y*=e,this.z*=e,this}multiplyVectors(e,t){return this.x=e.x*t.x,this.y=e.y*t.y,this.z=e.z*t.z,this}applyEuler(e){return this.applyQuaternion(Am.setFromEuler(e))}applyAxisAngle(e,t){return this.applyQuaternion(Am.setFromAxisAngle(e,t))}applyMatrix3(e){const t=this.x,s=this.y,a=this.z,l=e.elements;return this.x=l[0]*t+l[3]*s+l[6]*a,this.y=l[1]*t+l[4]*s+l[7]*a,this.z=l[2]*t+l[5]*s+l[8]*a,this}applyNormalMatrix(e){return this.applyMatrix3(e).normalize()}applyMatrix4(e){const t=this.x,s=this.y,a=this.z,l=e.elements,u=1/(l[3]*t+l[7]*s+l[11]*a+l[15]);return this.x=(l[0]*t+l[4]*s+l[8]*a+l[12])*u,this.y=(l[1]*t+l[5]*s+l[9]*a+l[13])*u,this.z=(l[2]*t+l[6]*s+l[10]*a+l[14])*u,this}applyQuaternion(e){const t=this.x,s=this.y,a=this.z,l=e.x,u=e.y,f=e.z,h=e.w,p=2*(u*a-f*s),m=2*(f*t-l*a),_=2*(l*s-u*t);return this.x=t+h*p+u*_-f*m,this.y=s+h*m+f*p-l*_,this.z=a+h*_+l*m-u*p,this}project(e){return this.applyMatrix4(e.matrixWorldInverse).applyMatrix4(e.projectionMatrix)}unproject(e){return this.applyMatrix4(e.projectionMatrixInverse).applyMatrix4(e.matrixWorld)}transformDirection(e){const t=this.x,s=this.y,a=this.z,l=e.elements;return this.x=l[0]*t+l[4]*s+l[8]*a,this.y=l[1]*t+l[5]*s+l[9]*a,this.z=l[2]*t+l[6]*s+l[10]*a,this.normalize()}divide(e){return this.x/=e.x,this.y/=e.y,this.z/=e.z,this}divideScalar(e){return this.multiplyScalar(1/e)}min(e){return this.x=Math.min(this.x,e.x),this.y=Math.min(this.y,e.y),this.z=Math.min(this.z,e.z),this}max(e){return this.x=Math.max(this.x,e.x),this.y=Math.max(this.y,e.y),this.z=Math.max(this.z,e.z),this}clamp(e,t){return this.x=Math.max(e.x,Math.min(t.x,this.x)),this.y=Math.max(e.y,Math.min(t.y,this.y)),this.z=Math.max(e.z,Math.min(t.z,this.z)),this}clampScalar(e,t){return this.x=Math.max(e,Math.min(t,this.x)),this.y=Math.max(e,Math.min(t,this.y)),this.z=Math.max(e,Math.min(t,this.z)),this}clampLength(e,t){const s=this.length();return this.divideScalar(s||1).multiplyScalar(Math.max(e,Math.min(t,s)))}floor(){return this.x=Math.floor(this.x),this.y=Math.floor(this.y),this.z=Math.floor(this.z),this}ceil(){return this.x=Math.ceil(this.x),this.y=Math.ceil(this.y),this.z=Math.ceil(this.z),this}round(){return this.x=Math.round(this.x),this.y=Math.round(this.y),this.z=Math.round(this.z),this}roundToZero(){return this.x=Math.trunc(this.x),this.y=Math.trunc(this.y),this.z=Math.trunc(this.z),this}negate(){return this.x=-this.x,this.y=-this.y,this.z=-this.z,this}dot(e){return this.x*e.x+this.y*e.y+this.z*e.z}lengthSq(){return this.x*this.x+this.y*this.y+this.z*this.z}length(){return Math.sqrt(this.x*this.x+this.y*this.y+this.z*this.z)}manhattanLength(){return Math.abs(this.x)+Math.abs(this.y)+Math.abs(this.z)}normalize(){return this.divideScalar(this.length()||1)}setLength(e){return this.normalize().multiplyScalar(e)}lerp(e,t){return this.x+=(e.x-this.x)*t,this.y+=(e.y-this.y)*t,this.z+=(e.z-this.z)*t,this}lerpVectors(e,t,s){return this.x=e.x+(t.x-e.x)*s,this.y=e.y+(t.y-e.y)*s,this.z=e.z+(t.z-e.z)*s,this}cross(e){return this.crossVectors(this,e)}crossVectors(e,t){const s=e.x,a=e.y,l=e.z,u=t.x,f=t.y,h=t.z;return this.x=a*h-l*f,this.y=l*u-s*h,this.z=s*f-a*u,this}projectOnVector(e){const t=e.lengthSq();if(t===0)return this.set(0,0,0);const s=e.dot(this)/t;return this.copy(e).multiplyScalar(s)}projectOnPlane(e){return Yu.copy(this).projectOnVector(e),this.sub(Yu)}reflect(e){return this.sub(Yu.copy(e).multiplyScalar(2*this.dot(e)))}angleTo(e){const t=Math.sqrt(this.lengthSq()*e.lengthSq());if(t===0)return Math.PI/2;const s=this.dot(e)/t;return Math.acos(Mn(s,-1,1))}distanceTo(e){return Math.sqrt(this.distanceToSquared(e))}distanceToSquared(e){const t=this.x-e.x,s=this.y-e.y,a=this.z-e.z;return t*t+s*s+a*a}manhattanDistanceTo(e){return Math.abs(this.x-e.x)+Math.abs(this.y-e.y)+Math.abs(this.z-e.z)}setFromSpherical(e){return this.setFromSphericalCoords(e.radius,e.phi,e.theta)}setFromSphericalCoords(e,t,s){const a=Math.sin(t)*e;return this.x=a*Math.sin(s),this.y=Math.cos(t)*e,this.z=a*Math.cos(s),this}setFromCylindrical(e){return this.setFromCylindricalCoords(e.radius,e.theta,e.y)}setFromCylindricalCoords(e,t,s){return this.x=e*Math.sin(t),this.y=s,this.z=e*Math.cos(t),this}setFromMatrixPosition(e){const t=e.elements;return this.x=t[12],this.y=t[13],this.z=t[14],this}setFromMatrixScale(e){const t=this.setFromMatrixColumn(e,0).length(),s=this.setFromMatrixColumn(e,1).length(),a=this.setFromMatrixColumn(e,2).length();return this.x=t,this.y=s,this.z=a,this}setFromMatrixColumn(e,t){return this.fromArray(e.elements,t*4)}setFromMatrix3Column(e,t){return this.fromArray(e.elements,t*3)}setFromEuler(e){return this.x=e._x,this.y=e._y,this.z=e._z,this}setFromColor(e){return this.x=e.r,this.y=e.g,this.z=e.b,this}equals(e){return e.x===this.x&&e.y===this.y&&e.z===this.z}fromArray(e,t=0){return this.x=e[t],this.y=e[t+1],this.z=e[t+2],this}toArray(e=[],t=0){return e[t]=this.x,e[t+1]=this.y,e[t+2]=this.z,e}fromBufferAttribute(e,t){return this.x=e.getX(t),this.y=e.getY(t),this.z=e.getZ(t),this}random(){return this.x=Math.random(),this.y=Math.random(),this.z=Math.random(),this}randomDirection(){const e=Math.random()*Math.PI*2,t=Math.random()*2-1,s=Math.sqrt(1-t*t);return this.x=s*Math.cos(e),this.y=t,this.z=s*Math.sin(e),this}*[Symbol.iterator](){yield this.x,yield this.y,yield this.z}}const Yu=new Q,Am=new rs;class ao{constructor(e=new Q(1/0,1/0,1/0),t=new Q(-1/0,-1/0,-1/0)){this.isBox3=!0,this.min=e,this.max=t}set(e,t){return this.min.copy(e),this.max.copy(t),this}setFromArray(e){this.makeEmpty();for(let t=0,s=e.length;t<s;t+=3)this.expandByPoint(oi.fromArray(e,t));return this}setFromBufferAttribute(e){this.makeEmpty();for(let t=0,s=e.count;t<s;t++)this.expandByPoint(oi.fromBufferAttribute(e,t));return this}setFromPoints(e){this.makeEmpty();for(let t=0,s=e.length;t<s;t++)this.expandByPoint(e[t]);return this}setFromCenterAndSize(e,t){const s=oi.copy(t).multiplyScalar(.5);return this.min.copy(e).sub(s),this.max.copy(e).add(s),this}setFromObject(e,t=!1){return this.makeEmpty(),this.expandByObject(e,t)}clone(){return new this.constructor().copy(this)}copy(e){return this.min.copy(e.min),this.max.copy(e.max),this}makeEmpty(){return this.min.x=this.min.y=this.min.z=1/0,this.max.x=this.max.y=this.max.z=-1/0,this}isEmpty(){return this.max.x<this.min.x||this.max.y<this.min.y||this.max.z<this.min.z}getCenter(e){return this.isEmpty()?e.set(0,0,0):e.addVectors(this.min,this.max).multiplyScalar(.5)}getSize(e){return this.isEmpty()?e.set(0,0,0):e.subVectors(this.max,this.min)}expandByPoint(e){return this.min.min(e),this.max.max(e),this}expandByVector(e){return this.min.sub(e),this.max.add(e),this}expandByScalar(e){return this.min.addScalar(-e),this.max.addScalar(e),this}expandByObject(e,t=!1){e.updateWorldMatrix(!1,!1);const s=e.geometry;if(s!==void 0){const l=s.getAttribute("position");if(t===!0&&l!==void 0&&e.isInstancedMesh!==!0)for(let u=0,f=l.count;u<f;u++)e.isMesh===!0?e.getVertexPosition(u,oi):oi.fromBufferAttribute(l,u),oi.applyMatrix4(e.matrixWorld),this.expandByPoint(oi);else e.boundingBox!==void 0?(e.boundingBox===null&&e.computeBoundingBox(),gl.copy(e.boundingBox)):(s.boundingBox===null&&s.computeBoundingBox(),gl.copy(s.boundingBox)),gl.applyMatrix4(e.matrixWorld),this.union(gl)}const a=e.children;for(let l=0,u=a.length;l<u;l++)this.expandByObject(a[l],t);return this}containsPoint(e){return e.x>=this.min.x&&e.x<=this.max.x&&e.y>=this.min.y&&e.y<=this.max.y&&e.z>=this.min.z&&e.z<=this.max.z}containsBox(e){return this.min.x<=e.min.x&&e.max.x<=this.max.x&&this.min.y<=e.min.y&&e.max.y<=this.max.y&&this.min.z<=e.min.z&&e.max.z<=this.max.z}getParameter(e,t){return t.set((e.x-this.min.x)/(this.max.x-this.min.x),(e.y-this.min.y)/(this.max.y-this.min.y),(e.z-this.min.z)/(this.max.z-this.min.z))}intersectsBox(e){return e.max.x>=this.min.x&&e.min.x<=this.max.x&&e.max.y>=this.min.y&&e.min.y<=this.max.y&&e.max.z>=this.min.z&&e.min.z<=this.max.z}intersectsSphere(e){return this.clampPoint(e.center,oi),oi.distanceToSquared(e.center)<=e.radius*e.radius}intersectsPlane(e){let t,s;return e.normal.x>0?(t=e.normal.x*this.min.x,s=e.normal.x*this.max.x):(t=e.normal.x*this.max.x,s=e.normal.x*this.min.x),e.normal.y>0?(t+=e.normal.y*this.min.y,s+=e.normal.y*this.max.y):(t+=e.normal.y*this.max.y,s+=e.normal.y*this.min.y),e.normal.z>0?(t+=e.normal.z*this.min.z,s+=e.normal.z*this.max.z):(t+=e.normal.z*this.max.z,s+=e.normal.z*this.min.z),t<=-e.constant&&s>=-e.constant}intersectsTriangle(e){if(this.isEmpty())return!1;this.getCenter(qo),_l.subVectors(this.max,qo),Fs.subVectors(e.a,qo),Os.subVectors(e.b,qo),ks.subVectors(e.c,qo),pr.subVectors(Os,Fs),mr.subVectors(ks,Os),Wr.subVectors(Fs,ks);let t=[0,-pr.z,pr.y,0,-mr.z,mr.y,0,-Wr.z,Wr.y,pr.z,0,-pr.x,mr.z,0,-mr.x,Wr.z,0,-Wr.x,-pr.y,pr.x,0,-mr.y,mr.x,0,-Wr.y,Wr.x,0];return!qu(t,Fs,Os,ks,_l)||(t=[1,0,0,0,1,0,0,0,1],!qu(t,Fs,Os,ks,_l))?!1:(vl.crossVectors(pr,mr),t=[vl.x,vl.y,vl.z],qu(t,Fs,Os,ks,_l))}clampPoint(e,t){return t.copy(e).clamp(this.min,this.max)}distanceToPoint(e){return this.clampPoint(e,oi).distanceTo(e)}getBoundingSphere(e){return this.isEmpty()?e.makeEmpty():(this.getCenter(e.center),e.radius=this.getSize(oi).length()*.5),e}intersect(e){return this.min.max(e.min),this.max.min(e.max),this.isEmpty()&&this.makeEmpty(),this}union(e){return this.min.min(e.min),this.max.max(e.max),this}applyMatrix4(e){return this.isEmpty()?this:(Ni[0].set(this.min.x,this.min.y,this.min.z).applyMatrix4(e),Ni[1].set(this.min.x,this.min.y,this.max.z).applyMatrix4(e),Ni[2].set(this.min.x,this.max.y,this.min.z).applyMatrix4(e),Ni[3].set(this.min.x,this.max.y,this.max.z).applyMatrix4(e),Ni[4].set(this.max.x,this.min.y,this.min.z).applyMatrix4(e),Ni[5].set(this.max.x,this.min.y,this.max.z).applyMatrix4(e),Ni[6].set(this.max.x,this.max.y,this.min.z).applyMatrix4(e),Ni[7].set(this.max.x,this.max.y,this.max.z).applyMatrix4(e),this.setFromPoints(Ni),this)}translate(e){return this.min.add(e),this.max.add(e),this}equals(e){return e.min.equals(this.min)&&e.max.equals(this.max)}}const Ni=[new Q,new Q,new Q,new Q,new Q,new Q,new Q,new Q],oi=new Q,gl=new ao,Fs=new Q,Os=new Q,ks=new Q,pr=new Q,mr=new Q,Wr=new Q,qo=new Q,_l=new Q,vl=new Q,jr=new Q;function qu(r,e,t,s,a){for(let l=0,u=r.length-3;l<=u;l+=3){jr.fromArray(r,l);const f=a.x*Math.abs(jr.x)+a.y*Math.abs(jr.y)+a.z*Math.abs(jr.z),h=e.dot(jr),p=t.dot(jr),m=s.dot(jr);if(Math.max(-Math.max(h,p,m),Math.min(h,p,m))>f)return!1}return!0}const py=new ao,$o=new Q,$u=new Q;class Td{constructor(e=new Q,t=-1){this.isSphere=!0,this.center=e,this.radius=t}set(e,t){return this.center.copy(e),this.radius=t,this}setFromPoints(e,t){const s=this.center;t!==void 0?s.copy(t):py.setFromPoints(e).getCenter(s);let a=0;for(let l=0,u=e.length;l<u;l++)a=Math.max(a,s.distanceToSquared(e[l]));return this.radius=Math.sqrt(a),this}copy(e){return this.center.copy(e.center),this.radius=e.radius,this}isEmpty(){return this.radius<0}makeEmpty(){return this.center.set(0,0,0),this.radius=-1,this}containsPoint(e){return e.distanceToSquared(this.center)<=this.radius*this.radius}distanceToPoint(e){return e.distanceTo(this.center)-this.radius}intersectsSphere(e){const t=this.radius+e.radius;return e.center.distanceToSquared(this.center)<=t*t}intersectsBox(e){return e.intersectsSphere(this)}intersectsPlane(e){return Math.abs(e.distanceToPoint(this.center))<=this.radius}clampPoint(e,t){const s=this.center.distanceToSquared(e);return t.copy(e),s>this.radius*this.radius&&(t.sub(this.center).normalize(),t.multiplyScalar(this.radius).add(this.center)),t}getBoundingBox(e){return this.isEmpty()?(e.makeEmpty(),e):(e.set(this.center,this.center),e.expandByScalar(this.radius),e)}applyMatrix4(e){return this.center.applyMatrix4(e),this.radius=this.radius*e.getMaxScaleOnAxis(),this}translate(e){return this.center.add(e),this}expandByPoint(e){if(this.isEmpty())return this.center.copy(e),this.radius=0,this;$o.subVectors(e,this.center);const t=$o.lengthSq();if(t>this.radius*this.radius){const s=Math.sqrt(t),a=(s-this.radius)*.5;this.center.addScaledVector($o,a/s),this.radius+=a}return this}union(e){return e.isEmpty()?this:this.isEmpty()?(this.copy(e),this):(this.center.equals(e.center)===!0?this.radius=Math.max(this.radius,e.radius):($u.subVectors(e.center,this.center).setLength(e.radius),this.expandByPoint($o.copy(e.center).add($u)),this.expandByPoint($o.copy(e.center).sub($u))),this)}equals(e){return e.center.equals(this.center)&&e.radius===this.radius}clone(){return new this.constructor().copy(this)}}const Ii=new Q,Ku=new Q,xl=new Q,gr=new Q,Zu=new Q,yl=new Q,Qu=new Q;class Kg{constructor(e=new Q,t=new Q(0,0,-1)){this.origin=e,this.direction=t}set(e,t){return this.origin.copy(e),this.direction.copy(t),this}copy(e){return this.origin.copy(e.origin),this.direction.copy(e.direction),this}at(e,t){return t.copy(this.origin).addScaledVector(this.direction,e)}lookAt(e){return this.direction.copy(e).sub(this.origin).normalize(),this}recast(e){return this.origin.copy(this.at(e,Ii)),this}closestPointToPoint(e,t){t.subVectors(e,this.origin);const s=t.dot(this.direction);return s<0?t.copy(this.origin):t.copy(this.origin).addScaledVector(this.direction,s)}distanceToPoint(e){return Math.sqrt(this.distanceSqToPoint(e))}distanceSqToPoint(e){const t=Ii.subVectors(e,this.origin).dot(this.direction);return t<0?this.origin.distanceToSquared(e):(Ii.copy(this.origin).addScaledVector(this.direction,t),Ii.distanceToSquared(e))}distanceSqToSegment(e,t,s,a){Ku.copy(e).add(t).multiplyScalar(.5),xl.copy(t).sub(e).normalize(),gr.copy(this.origin).sub(Ku);const l=e.distanceTo(t)*.5,u=-this.direction.dot(xl),f=gr.dot(this.direction),h=-gr.dot(xl),p=gr.lengthSq(),m=Math.abs(1-u*u);let _,x,S,T;if(m>0)if(_=u*h-f,x=u*f-h,T=l*m,_>=0)if(x>=-T)if(x<=T){const w=1/m;_*=w,x*=w,S=_*(_+u*x+2*f)+x*(u*_+x+2*h)+p}else x=l,_=Math.max(0,-(u*x+f)),S=-_*_+x*(x+2*h)+p;else x=-l,_=Math.max(0,-(u*x+f)),S=-_*_+x*(x+2*h)+p;else x<=-T?(_=Math.max(0,-(-u*l+f)),x=_>0?-l:Math.min(Math.max(-l,-h),l),S=-_*_+x*(x+2*h)+p):x<=T?(_=0,x=Math.min(Math.max(-l,-h),l),S=x*(x+2*h)+p):(_=Math.max(0,-(u*l+f)),x=_>0?l:Math.min(Math.max(-l,-h),l),S=-_*_+x*(x+2*h)+p);else x=u>0?-l:l,_=Math.max(0,-(u*x+f)),S=-_*_+x*(x+2*h)+p;return s&&s.copy(this.origin).addScaledVector(this.direction,_),a&&a.copy(Ku).addScaledVector(xl,x),S}intersectSphere(e,t){Ii.subVectors(e.center,this.origin);const s=Ii.dot(this.direction),a=Ii.dot(Ii)-s*s,l=e.radius*e.radius;if(a>l)return null;const u=Math.sqrt(l-a),f=s-u,h=s+u;return h<0?null:f<0?this.at(h,t):this.at(f,t)}intersectsSphere(e){return this.distanceSqToPoint(e.center)<=e.radius*e.radius}distanceToPlane(e){const t=e.normal.dot(this.direction);if(t===0)return e.distanceToPoint(this.origin)===0?0:null;const s=-(this.origin.dot(e.normal)+e.constant)/t;return s>=0?s:null}intersectPlane(e,t){const s=this.distanceToPlane(e);return s===null?null:this.at(s,t)}intersectsPlane(e){const t=e.distanceToPoint(this.origin);return t===0||e.normal.dot(this.direction)*t<0}intersectBox(e,t){let s,a,l,u,f,h;const p=1/this.direction.x,m=1/this.direction.y,_=1/this.direction.z,x=this.origin;return p>=0?(s=(e.min.x-x.x)*p,a=(e.max.x-x.x)*p):(s=(e.max.x-x.x)*p,a=(e.min.x-x.x)*p),m>=0?(l=(e.min.y-x.y)*m,u=(e.max.y-x.y)*m):(l=(e.max.y-x.y)*m,u=(e.min.y-x.y)*m),s>u||l>a||((l>s||isNaN(s))&&(s=l),(u<a||isNaN(a))&&(a=u),_>=0?(f=(e.min.z-x.z)*_,h=(e.max.z-x.z)*_):(f=(e.max.z-x.z)*_,h=(e.min.z-x.z)*_),s>h||f>a)||((f>s||s!==s)&&(s=f),(h<a||a!==a)&&(a=h),a<0)?null:this.at(s>=0?s:a,t)}intersectsBox(e){return this.intersectBox(e,Ii)!==null}intersectTriangle(e,t,s,a,l){Zu.subVectors(t,e),yl.subVectors(s,e),Qu.crossVectors(Zu,yl);let u=this.direction.dot(Qu),f;if(u>0){if(a)return null;f=1}else if(u<0)f=-1,u=-u;else return null;gr.subVectors(this.origin,e);const h=f*this.direction.dot(yl.crossVectors(gr,yl));if(h<0)return null;const p=f*this.direction.dot(Zu.cross(gr));if(p<0||h+p>u)return null;const m=-f*gr.dot(Qu);return m<0?null:this.at(m/u,l)}applyMatrix4(e){return this.origin.applyMatrix4(e),this.direction.transformDirection(e),this}equals(e){return e.origin.equals(this.origin)&&e.direction.equals(this.direction)}clone(){return new this.constructor().copy(this)}}class Gt{constructor(e,t,s,a,l,u,f,h,p,m,_,x,S,T,w,v){Gt.prototype.isMatrix4=!0,this.elements=[1,0,0,0,0,1,0,0,0,0,1,0,0,0,0,1],e!==void 0&&this.set(e,t,s,a,l,u,f,h,p,m,_,x,S,T,w,v)}set(e,t,s,a,l,u,f,h,p,m,_,x,S,T,w,v){const y=this.elements;return y[0]=e,y[4]=t,y[8]=s,y[12]=a,y[1]=l,y[5]=u,y[9]=f,y[13]=h,y[2]=p,y[6]=m,y[10]=_,y[14]=x,y[3]=S,y[7]=T,y[11]=w,y[15]=v,this}identity(){return this.set(1,0,0,0,0,1,0,0,0,0,1,0,0,0,0,1),this}clone(){return new Gt().fromArray(this.elements)}copy(e){const t=this.elements,s=e.elements;return t[0]=s[0],t[1]=s[1],t[2]=s[2],t[3]=s[3],t[4]=s[4],t[5]=s[5],t[6]=s[6],t[7]=s[7],t[8]=s[8],t[9]=s[9],t[10]=s[10],t[11]=s[11],t[12]=s[12],t[13]=s[13],t[14]=s[14],t[15]=s[15],this}copyPosition(e){const t=this.elements,s=e.elements;return t[12]=s[12],t[13]=s[13],t[14]=s[14],this}setFromMatrix3(e){const t=e.elements;return this.set(t[0],t[3],t[6],0,t[1],t[4],t[7],0,t[2],t[5],t[8],0,0,0,0,1),this}extractBasis(e,t,s){return e.setFromMatrixColumn(this,0),t.setFromMatrixColumn(this,1),s.setFromMatrixColumn(this,2),this}makeBasis(e,t,s){return this.set(e.x,t.x,s.x,0,e.y,t.y,s.y,0,e.z,t.z,s.z,0,0,0,0,1),this}extractRotation(e){const t=this.elements,s=e.elements,a=1/Bs.setFromMatrixColumn(e,0).length(),l=1/Bs.setFromMatrixColumn(e,1).length(),u=1/Bs.setFromMatrixColumn(e,2).length();return t[0]=s[0]*a,t[1]=s[1]*a,t[2]=s[2]*a,t[3]=0,t[4]=s[4]*l,t[5]=s[5]*l,t[6]=s[6]*l,t[7]=0,t[8]=s[8]*u,t[9]=s[9]*u,t[10]=s[10]*u,t[11]=0,t[12]=0,t[13]=0,t[14]=0,t[15]=1,this}makeRotationFromEuler(e){const t=this.elements,s=e.x,a=e.y,l=e.z,u=Math.cos(s),f=Math.sin(s),h=Math.cos(a),p=Math.sin(a),m=Math.cos(l),_=Math.sin(l);if(e.order==="XYZ"){const x=u*m,S=u*_,T=f*m,w=f*_;t[0]=h*m,t[4]=-h*_,t[8]=p,t[1]=S+T*p,t[5]=x-w*p,t[9]=-f*h,t[2]=w-x*p,t[6]=T+S*p,t[10]=u*h}else if(e.order==="YXZ"){const x=h*m,S=h*_,T=p*m,w=p*_;t[0]=x+w*f,t[4]=T*f-S,t[8]=u*p,t[1]=u*_,t[5]=u*m,t[9]=-f,t[2]=S*f-T,t[6]=w+x*f,t[10]=u*h}else if(e.order==="ZXY"){const x=h*m,S=h*_,T=p*m,w=p*_;t[0]=x-w*f,t[4]=-u*_,t[8]=T+S*f,t[1]=S+T*f,t[5]=u*m,t[9]=w-x*f,t[2]=-u*p,t[6]=f,t[10]=u*h}else if(e.order==="ZYX"){const x=u*m,S=u*_,T=f*m,w=f*_;t[0]=h*m,t[4]=T*p-S,t[8]=x*p+w,t[1]=h*_,t[5]=w*p+x,t[9]=S*p-T,t[2]=-p,t[6]=f*h,t[10]=u*h}else if(e.order==="YZX"){const x=u*h,S=u*p,T=f*h,w=f*p;t[0]=h*m,t[4]=w-x*_,t[8]=T*_+S,t[1]=_,t[5]=u*m,t[9]=-f*m,t[2]=-p*m,t[6]=S*_+T,t[10]=x-w*_}else if(e.order==="XZY"){const x=u*h,S=u*p,T=f*h,w=f*p;t[0]=h*m,t[4]=-_,t[8]=p*m,t[1]=x*_+w,t[5]=u*m,t[9]=S*_-T,t[2]=T*_-S,t[6]=f*m,t[10]=w*_+x}return t[3]=0,t[7]=0,t[11]=0,t[12]=0,t[13]=0,t[14]=0,t[15]=1,this}makeRotationFromQuaternion(e){return this.compose(my,e,gy)}lookAt(e,t,s){const a=this.elements;return zn.subVectors(e,t),zn.lengthSq()===0&&(zn.z=1),zn.normalize(),_r.crossVectors(s,zn),_r.lengthSq()===0&&(Math.abs(s.z)===1?zn.x+=1e-4:zn.z+=1e-4,zn.normalize(),_r.crossVectors(s,zn)),_r.normalize(),Sl.crossVectors(zn,_r),a[0]=_r.x,a[4]=Sl.x,a[8]=zn.x,a[1]=_r.y,a[5]=Sl.y,a[9]=zn.y,a[2]=_r.z,a[6]=Sl.z,a[10]=zn.z,this}multiply(e){return this.multiplyMatrices(this,e)}premultiply(e){return this.multiplyMatrices(e,this)}multiplyMatrices(e,t){const s=e.elements,a=t.elements,l=this.elements,u=s[0],f=s[4],h=s[8],p=s[12],m=s[1],_=s[5],x=s[9],S=s[13],T=s[2],w=s[6],v=s[10],y=s[14],P=s[3],b=s[7],D=s[11],V=s[15],O=a[0],U=a[4],Y=a[8],ce=a[12],E=a[1],C=a[5],re=a[9],ee=a[13],ae=a[2],ue=a[6],Z=a[10],le=a[14],F=a[3],se=a[7],L=a[11],X=a[15];return l[0]=u*O+f*E+h*ae+p*F,l[4]=u*U+f*C+h*ue+p*se,l[8]=u*Y+f*re+h*Z+p*L,l[12]=u*ce+f*ee+h*le+p*X,l[1]=m*O+_*E+x*ae+S*F,l[5]=m*U+_*C+x*ue+S*se,l[9]=m*Y+_*re+x*Z+S*L,l[13]=m*ce+_*ee+x*le+S*X,l[2]=T*O+w*E+v*ae+y*F,l[6]=T*U+w*C+v*ue+y*se,l[10]=T*Y+w*re+v*Z+y*L,l[14]=T*ce+w*ee+v*le+y*X,l[3]=P*O+b*E+D*ae+V*F,l[7]=P*U+b*C+D*ue+V*se,l[11]=P*Y+b*re+D*Z+V*L,l[15]=P*ce+b*ee+D*le+V*X,this}multiplyScalar(e){const t=this.elements;return t[0]*=e,t[4]*=e,t[8]*=e,t[12]*=e,t[1]*=e,t[5]*=e,t[9]*=e,t[13]*=e,t[2]*=e,t[6]*=e,t[10]*=e,t[14]*=e,t[3]*=e,t[7]*=e,t[11]*=e,t[15]*=e,this}determinant(){const e=this.elements,t=e[0],s=e[4],a=e[8],l=e[12],u=e[1],f=e[5],h=e[9],p=e[13],m=e[2],_=e[6],x=e[10],S=e[14],T=e[3],w=e[7],v=e[11],y=e[15];return T*(+l*h*_-a*p*_-l*f*x+s*p*x+a*f*S-s*h*S)+w*(+t*h*S-t*p*x+l*u*x-a*u*S+a*p*m-l*h*m)+v*(+t*p*_-t*f*S-l*u*_+s*u*S+l*f*m-s*p*m)+y*(-a*f*m-t*h*_+t*f*x+a*u*_-s*u*x+s*h*m)}transpose(){const e=this.elements;let t;return t=e[1],e[1]=e[4],e[4]=t,t=e[2],e[2]=e[8],e[8]=t,t=e[6],e[6]=e[9],e[9]=t,t=e[3],e[3]=e[12],e[12]=t,t=e[7],e[7]=e[13],e[13]=t,t=e[11],e[11]=e[14],e[14]=t,this}setPosition(e,t,s){const a=this.elements;return e.isVector3?(a[12]=e.x,a[13]=e.y,a[14]=e.z):(a[12]=e,a[13]=t,a[14]=s),this}invert(){const e=this.elements,t=e[0],s=e[1],a=e[2],l=e[3],u=e[4],f=e[5],h=e[6],p=e[7],m=e[8],_=e[9],x=e[10],S=e[11],T=e[12],w=e[13],v=e[14],y=e[15],P=_*v*p-w*x*p+w*h*S-f*v*S-_*h*y+f*x*y,b=T*x*p-m*v*p-T*h*S+u*v*S+m*h*y-u*x*y,D=m*w*p-T*_*p+T*f*S-u*w*S-m*f*y+u*_*y,V=T*_*h-m*w*h-T*f*x+u*w*x+m*f*v-u*_*v,O=t*P+s*b+a*D+l*V;if(O===0)return this.set(0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0);const U=1/O;return e[0]=P*U,e[1]=(w*x*l-_*v*l-w*a*S+s*v*S+_*a*y-s*x*y)*U,e[2]=(f*v*l-w*h*l+w*a*p-s*v*p-f*a*y+s*h*y)*U,e[3]=(_*h*l-f*x*l-_*a*p+s*x*p+f*a*S-s*h*S)*U,e[4]=b*U,e[5]=(m*v*l-T*x*l+T*a*S-t*v*S-m*a*y+t*x*y)*U,e[6]=(T*h*l-u*v*l-T*a*p+t*v*p+u*a*y-t*h*y)*U,e[7]=(u*x*l-m*h*l+m*a*p-t*x*p-u*a*S+t*h*S)*U,e[8]=D*U,e[9]=(T*_*l-m*w*l-T*s*S+t*w*S+m*s*y-t*_*y)*U,e[10]=(u*w*l-T*f*l+T*s*p-t*w*p-u*s*y+t*f*y)*U,e[11]=(m*f*l-u*_*l-m*s*p+t*_*p+u*s*S-t*f*S)*U,e[12]=V*U,e[13]=(m*w*a-T*_*a+T*s*x-t*w*x-m*s*v+t*_*v)*U,e[14]=(T*f*a-u*w*a-T*s*h+t*w*h+u*s*v-t*f*v)*U,e[15]=(u*_*a-m*f*a+m*s*h-t*_*h-u*s*x+t*f*x)*U,this}scale(e){const t=this.elements,s=e.x,a=e.y,l=e.z;return t[0]*=s,t[4]*=a,t[8]*=l,t[1]*=s,t[5]*=a,t[9]*=l,t[2]*=s,t[6]*=a,t[10]*=l,t[3]*=s,t[7]*=a,t[11]*=l,this}getMaxScaleOnAxis(){const e=this.elements,t=e[0]*e[0]+e[1]*e[1]+e[2]*e[2],s=e[4]*e[4]+e[5]*e[5]+e[6]*e[6],a=e[8]*e[8]+e[9]*e[9]+e[10]*e[10];return Math.sqrt(Math.max(t,s,a))}makeTranslation(e,t,s){return e.isVector3?this.set(1,0,0,e.x,0,1,0,e.y,0,0,1,e.z,0,0,0,1):this.set(1,0,0,e,0,1,0,t,0,0,1,s,0,0,0,1),this}makeRotationX(e){const t=Math.cos(e),s=Math.sin(e);return this.set(1,0,0,0,0,t,-s,0,0,s,t,0,0,0,0,1),this}makeRotationY(e){const t=Math.cos(e),s=Math.sin(e);return this.set(t,0,s,0,0,1,0,0,-s,0,t,0,0,0,0,1),this}makeRotationZ(e){const t=Math.cos(e),s=Math.sin(e);return this.set(t,-s,0,0,s,t,0,0,0,0,1,0,0,0,0,1),this}makeRotationAxis(e,t){const s=Math.cos(t),a=Math.sin(t),l=1-s,u=e.x,f=e.y,h=e.z,p=l*u,m=l*f;return this.set(p*u+s,p*f-a*h,p*h+a*f,0,p*f+a*h,m*f+s,m*h-a*u,0,p*h-a*f,m*h+a*u,l*h*h+s,0,0,0,0,1),this}makeScale(e,t,s){return this.set(e,0,0,0,0,t,0,0,0,0,s,0,0,0,0,1),this}makeShear(e,t,s,a,l,u){return this.set(1,s,l,0,e,1,u,0,t,a,1,0,0,0,0,1),this}compose(e,t,s){const a=this.elements,l=t._x,u=t._y,f=t._z,h=t._w,p=l+l,m=u+u,_=f+f,x=l*p,S=l*m,T=l*_,w=u*m,v=u*_,y=f*_,P=h*p,b=h*m,D=h*_,V=s.x,O=s.y,U=s.z;return a[0]=(1-(w+y))*V,a[1]=(S+D)*V,a[2]=(T-b)*V,a[3]=0,a[4]=(S-D)*O,a[5]=(1-(x+y))*O,a[6]=(v+P)*O,a[7]=0,a[8]=(T+b)*U,a[9]=(v-P)*U,a[10]=(1-(x+w))*U,a[11]=0,a[12]=e.x,a[13]=e.y,a[14]=e.z,a[15]=1,this}decompose(e,t,s){const a=this.elements;let l=Bs.set(a[0],a[1],a[2]).length();const u=Bs.set(a[4],a[5],a[6]).length(),f=Bs.set(a[8],a[9],a[10]).length();this.determinant()<0&&(l=-l),e.x=a[12],e.y=a[13],e.z=a[14],ai.copy(this);const p=1/l,m=1/u,_=1/f;return ai.elements[0]*=p,ai.elements[1]*=p,ai.elements[2]*=p,ai.elements[4]*=m,ai.elements[5]*=m,ai.elements[6]*=m,ai.elements[8]*=_,ai.elements[9]*=_,ai.elements[10]*=_,t.setFromRotationMatrix(ai),s.x=l,s.y=u,s.z=f,this}makePerspective(e,t,s,a,l,u,f=Vi){const h=this.elements,p=2*l/(t-e),m=2*l/(s-a),_=(t+e)/(t-e),x=(s+a)/(s-a);let S,T;if(f===Vi)S=-(u+l)/(u-l),T=-2*u*l/(u-l);else if(f===Kl)S=-u/(u-l),T=-u*l/(u-l);else throw new Error("THREE.Matrix4.makePerspective(): Invalid coordinate system: "+f);return h[0]=p,h[4]=0,h[8]=_,h[12]=0,h[1]=0,h[5]=m,h[9]=x,h[13]=0,h[2]=0,h[6]=0,h[10]=S,h[14]=T,h[3]=0,h[7]=0,h[11]=-1,h[15]=0,this}makeOrthographic(e,t,s,a,l,u,f=Vi){const h=this.elements,p=1/(t-e),m=1/(s-a),_=1/(u-l),x=(t+e)*p,S=(s+a)*m;let T,w;if(f===Vi)T=(u+l)*_,w=-2*_;else if(f===Kl)T=l*_,w=-1*_;else throw new Error("THREE.Matrix4.makeOrthographic(): Invalid coordinate system: "+f);return h[0]=2*p,h[4]=0,h[8]=0,h[12]=-x,h[1]=0,h[5]=2*m,h[9]=0,h[13]=-S,h[2]=0,h[6]=0,h[10]=w,h[14]=-T,h[3]=0,h[7]=0,h[11]=0,h[15]=1,this}equals(e){const t=this.elements,s=e.elements;for(let a=0;a<16;a++)if(t[a]!==s[a])return!1;return!0}fromArray(e,t=0){for(let s=0;s<16;s++)this.elements[s]=e[s+t];return this}toArray(e=[],t=0){const s=this.elements;return e[t]=s[0],e[t+1]=s[1],e[t+2]=s[2],e[t+3]=s[3],e[t+4]=s[4],e[t+5]=s[5],e[t+6]=s[6],e[t+7]=s[7],e[t+8]=s[8],e[t+9]=s[9],e[t+10]=s[10],e[t+11]=s[11],e[t+12]=s[12],e[t+13]=s[13],e[t+14]=s[14],e[t+15]=s[15],e}}const Bs=new Q,ai=new Gt,my=new Q(0,0,0),gy=new Q(1,1,1),_r=new Q,Sl=new Q,zn=new Q,Cm=new Gt,Rm=new rs;class Si{constructor(e=0,t=0,s=0,a=Si.DEFAULT_ORDER){this.isEuler=!0,this._x=e,this._y=t,this._z=s,this._order=a}get x(){return this._x}set x(e){this._x=e,this._onChangeCallback()}get y(){return this._y}set y(e){this._y=e,this._onChangeCallback()}get z(){return this._z}set z(e){this._z=e,this._onChangeCallback()}get order(){return this._order}set order(e){this._order=e,this._onChangeCallback()}set(e,t,s,a=this._order){return this._x=e,this._y=t,this._z=s,this._order=a,this._onChangeCallback(),this}clone(){return new this.constructor(this._x,this._y,this._z,this._order)}copy(e){return this._x=e._x,this._y=e._y,this._z=e._z,this._order=e._order,this._onChangeCallback(),this}setFromRotationMatrix(e,t=this._order,s=!0){const a=e.elements,l=a[0],u=a[4],f=a[8],h=a[1],p=a[5],m=a[9],_=a[2],x=a[6],S=a[10];switch(t){case"XYZ":this._y=Math.asin(Mn(f,-1,1)),Math.abs(f)<.9999999?(this._x=Math.atan2(-m,S),this._z=Math.atan2(-u,l)):(this._x=Math.atan2(x,p),this._z=0);break;case"YXZ":this._x=Math.asin(-Mn(m,-1,1)),Math.abs(m)<.9999999?(this._y=Math.atan2(f,S),this._z=Math.atan2(h,p)):(this._y=Math.atan2(-_,l),this._z=0);break;case"ZXY":this._x=Math.asin(Mn(x,-1,1)),Math.abs(x)<.9999999?(this._y=Math.atan2(-_,S),this._z=Math.atan2(-u,p)):(this._y=0,this._z=Math.atan2(h,l));break;case"ZYX":this._y=Math.asin(-Mn(_,-1,1)),Math.abs(_)<.9999999?(this._x=Math.atan2(x,S),this._z=Math.atan2(h,l)):(this._x=0,this._z=Math.atan2(-u,p));break;case"YZX":this._z=Math.asin(Mn(h,-1,1)),Math.abs(h)<.9999999?(this._x=Math.atan2(-m,p),this._y=Math.atan2(-_,l)):(this._x=0,this._y=Math.atan2(f,S));break;case"XZY":this._z=Math.asin(-Mn(u,-1,1)),Math.abs(u)<.9999999?(this._x=Math.atan2(x,p),this._y=Math.atan2(f,l)):(this._x=Math.atan2(-m,S),this._y=0);break;default:console.warn("THREE.Euler: .setFromRotationMatrix() encountered an unknown order: "+t)}return this._order=t,s===!0&&this._onChangeCallback(),this}setFromQuaternion(e,t,s){return Cm.makeRotationFromQuaternion(e),this.setFromRotationMatrix(Cm,t,s)}setFromVector3(e,t=this._order){return this.set(e.x,e.y,e.z,t)}reorder(e){return Rm.setFromEuler(this),this.setFromQuaternion(Rm,e)}equals(e){return e._x===this._x&&e._y===this._y&&e._z===this._z&&e._order===this._order}fromArray(e){return this._x=e[0],this._y=e[1],this._z=e[2],e[3]!==void 0&&(this._order=e[3]),this._onChangeCallback(),this}toArray(e=[],t=0){return e[t]=this._x,e[t+1]=this._y,e[t+2]=this._z,e[t+3]=this._order,e}_onChange(e){return this._onChangeCallback=e,this}_onChangeCallback(){}*[Symbol.iterator](){yield this._x,yield this._y,yield this._z,yield this._order}}Si.DEFAULT_ORDER="XYZ";class Zg{constructor(){this.mask=1}set(e){this.mask=(1<<e|0)>>>0}enable(e){this.mask|=1<<e|0}enableAll(){this.mask=-1}toggle(e){this.mask^=1<<e|0}disable(e){this.mask&=~(1<<e|0)}disableAll(){this.mask=0}test(e){return(this.mask&e.mask)!==0}isEnabled(e){return(this.mask&(1<<e|0))!==0}}let _y=0;const bm=new Q,zs=new rs,Ui=new Gt,Ml=new Q,Ko=new Q,vy=new Q,xy=new rs,Pm=new Q(1,0,0),Lm=new Q(0,1,0),Dm=new Q(0,0,1),Nm={type:"added"},yy={type:"removed"},Hs={type:"childadded",child:null},Ju={type:"childremoved",child:null};class vn extends os{constructor(){super(),this.isObject3D=!0,Object.defineProperty(this,"id",{value:_y++}),this.uuid=oa(),this.name="",this.type="Object3D",this.parent=null,this.children=[],this.up=vn.DEFAULT_UP.clone();const e=new Q,t=new Si,s=new rs,a=new Q(1,1,1);function l(){s.setFromEuler(t,!1)}function u(){t.setFromQuaternion(s,void 0,!1)}t._onChange(l),s._onChange(u),Object.defineProperties(this,{position:{configurable:!0,enumerable:!0,value:e},rotation:{configurable:!0,enumerable:!0,value:t},quaternion:{configurable:!0,enumerable:!0,value:s},scale:{configurable:!0,enumerable:!0,value:a},modelViewMatrix:{value:new Gt},normalMatrix:{value:new ot}}),this.matrix=new Gt,this.matrixWorld=new Gt,this.matrixAutoUpdate=vn.DEFAULT_MATRIX_AUTO_UPDATE,this.matrixWorldAutoUpdate=vn.DEFAULT_MATRIX_WORLD_AUTO_UPDATE,this.matrixWorldNeedsUpdate=!1,this.layers=new Zg,this.visible=!0,this.castShadow=!1,this.receiveShadow=!1,this.frustumCulled=!0,this.renderOrder=0,this.animations=[],this.userData={}}onBeforeShadow(){}onAfterShadow(){}onBeforeRender(){}onAfterRender(){}applyMatrix4(e){this.matrixAutoUpdate&&this.updateMatrix(),this.matrix.premultiply(e),this.matrix.decompose(this.position,this.quaternion,this.scale)}applyQuaternion(e){return this.quaternion.premultiply(e),this}setRotationFromAxisAngle(e,t){this.quaternion.setFromAxisAngle(e,t)}setRotationFromEuler(e){this.quaternion.setFromEuler(e,!0)}setRotationFromMatrix(e){this.quaternion.setFromRotationMatrix(e)}setRotationFromQuaternion(e){this.quaternion.copy(e)}rotateOnAxis(e,t){return zs.setFromAxisAngle(e,t),this.quaternion.multiply(zs),this}rotateOnWorldAxis(e,t){return zs.setFromAxisAngle(e,t),this.quaternion.premultiply(zs),this}rotateX(e){return this.rotateOnAxis(Pm,e)}rotateY(e){return this.rotateOnAxis(Lm,e)}rotateZ(e){return this.rotateOnAxis(Dm,e)}translateOnAxis(e,t){return bm.copy(e).applyQuaternion(this.quaternion),this.position.add(bm.multiplyScalar(t)),this}translateX(e){return this.translateOnAxis(Pm,e)}translateY(e){return this.translateOnAxis(Lm,e)}translateZ(e){return this.translateOnAxis(Dm,e)}localToWorld(e){return this.updateWorldMatrix(!0,!1),e.applyMatrix4(this.matrixWorld)}worldToLocal(e){return this.updateWorldMatrix(!0,!1),e.applyMatrix4(Ui.copy(this.matrixWorld).invert())}lookAt(e,t,s){e.isVector3?Ml.copy(e):Ml.set(e,t,s);const a=this.parent;this.updateWorldMatrix(!0,!1),Ko.setFromMatrixPosition(this.matrixWorld),this.isCamera||this.isLight?Ui.lookAt(Ko,Ml,this.up):Ui.lookAt(Ml,Ko,this.up),this.quaternion.setFromRotationMatrix(Ui),a&&(Ui.extractRotation(a.matrixWorld),zs.setFromRotationMatrix(Ui),this.quaternion.premultiply(zs.invert()))}add(e){if(arguments.length>1){for(let t=0;t<arguments.length;t++)this.add(arguments[t]);return this}return e===this?(console.error("THREE.Object3D.add: object can't be added as a child of itself.",e),this):(e&&e.isObject3D?(e.removeFromParent(),e.parent=this,this.children.push(e),e.dispatchEvent(Nm),Hs.child=e,this.dispatchEvent(Hs),Hs.child=null):console.error("THREE.Object3D.add: object not an instance of THREE.Object3D.",e),this)}remove(e){if(arguments.length>1){for(let s=0;s<arguments.length;s++)this.remove(arguments[s]);return this}const t=this.children.indexOf(e);return t!==-1&&(e.parent=null,this.children.splice(t,1),e.dispatchEvent(yy),Ju.child=e,this.dispatchEvent(Ju),Ju.child=null),this}removeFromParent(){const e=this.parent;return e!==null&&e.remove(this),this}clear(){return this.remove(...this.children)}attach(e){return this.updateWorldMatrix(!0,!1),Ui.copy(this.matrixWorld).invert(),e.parent!==null&&(e.parent.updateWorldMatrix(!0,!1),Ui.multiply(e.parent.matrixWorld)),e.applyMatrix4(Ui),e.removeFromParent(),e.parent=this,this.children.push(e),e.updateWorldMatrix(!1,!0),e.dispatchEvent(Nm),Hs.child=e,this.dispatchEvent(Hs),Hs.child=null,this}getObjectById(e){return this.getObjectByProperty("id",e)}getObjectByName(e){return this.getObjectByProperty("name",e)}getObjectByProperty(e,t){if(this[e]===t)return this;for(let s=0,a=this.children.length;s<a;s++){const u=this.children[s].getObjectByProperty(e,t);if(u!==void 0)return u}}getObjectsByProperty(e,t,s=[]){this[e]===t&&s.push(this);const a=this.children;for(let l=0,u=a.length;l<u;l++)a[l].getObjectsByProperty(e,t,s);return s}getWorldPosition(e){return this.updateWorldMatrix(!0,!1),e.setFromMatrixPosition(this.matrixWorld)}getWorldQuaternion(e){return this.updateWorldMatrix(!0,!1),this.matrixWorld.decompose(Ko,e,vy),e}getWorldScale(e){return this.updateWorldMatrix(!0,!1),this.matrixWorld.decompose(Ko,xy,e),e}getWorldDirection(e){this.updateWorldMatrix(!0,!1);const t=this.matrixWorld.elements;return e.set(t[8],t[9],t[10]).normalize()}raycast(){}traverse(e){e(this);const t=this.children;for(let s=0,a=t.length;s<a;s++)t[s].traverse(e)}traverseVisible(e){if(this.visible===!1)return;e(this);const t=this.children;for(let s=0,a=t.length;s<a;s++)t[s].traverseVisible(e)}traverseAncestors(e){const t=this.parent;t!==null&&(e(t),t.traverseAncestors(e))}updateMatrix(){this.matrix.compose(this.position,this.quaternion,this.scale),this.matrixWorldNeedsUpdate=!0}updateMatrixWorld(e){this.matrixAutoUpdate&&this.updateMatrix(),(this.matrixWorldNeedsUpdate||e)&&(this.matrixWorldAutoUpdate===!0&&(this.parent===null?this.matrixWorld.copy(this.matrix):this.matrixWorld.multiplyMatrices(this.parent.matrixWorld,this.matrix)),this.matrixWorldNeedsUpdate=!1,e=!0);const t=this.children;for(let s=0,a=t.length;s<a;s++)t[s].updateMatrixWorld(e)}updateWorldMatrix(e,t){const s=this.parent;if(e===!0&&s!==null&&s.updateWorldMatrix(!0,!1),this.matrixAutoUpdate&&this.updateMatrix(),this.matrixWorldAutoUpdate===!0&&(this.parent===null?this.matrixWorld.copy(this.matrix):this.matrixWorld.multiplyMatrices(this.parent.matrixWorld,this.matrix)),t===!0){const a=this.children;for(let l=0,u=a.length;l<u;l++)a[l].updateWorldMatrix(!1,!0)}}toJSON(e){const t=e===void 0||typeof e=="string",s={};t&&(e={geometries:{},materials:{},textures:{},images:{},shapes:{},skeletons:{},animations:{},nodes:{}},s.metadata={version:4.6,type:"Object",generator:"Object3D.toJSON"});const a={};a.uuid=this.uuid,a.type=this.type,this.name!==""&&(a.name=this.name),this.castShadow===!0&&(a.castShadow=!0),this.receiveShadow===!0&&(a.receiveShadow=!0),this.visible===!1&&(a.visible=!1),this.frustumCulled===!1&&(a.frustumCulled=!1),this.renderOrder!==0&&(a.renderOrder=this.renderOrder),Object.keys(this.userData).length>0&&(a.userData=this.userData),a.layers=this.layers.mask,a.matrix=this.matrix.toArray(),a.up=this.up.toArray(),this.matrixAutoUpdate===!1&&(a.matrixAutoUpdate=!1),this.isInstancedMesh&&(a.type="InstancedMesh",a.count=this.count,a.instanceMatrix=this.instanceMatrix.toJSON(),this.instanceColor!==null&&(a.instanceColor=this.instanceColor.toJSON())),this.isBatchedMesh&&(a.type="BatchedMesh",a.perObjectFrustumCulled=this.perObjectFrustumCulled,a.sortObjects=this.sortObjects,a.drawRanges=this._drawRanges,a.reservedRanges=this._reservedRanges,a.visibility=this._visibility,a.active=this._active,a.bounds=this._bounds.map(f=>({boxInitialized:f.boxInitialized,boxMin:f.box.min.toArray(),boxMax:f.box.max.toArray(),sphereInitialized:f.sphereInitialized,sphereRadius:f.sphere.radius,sphereCenter:f.sphere.center.toArray()})),a.maxInstanceCount=this._maxInstanceCount,a.maxVertexCount=this._maxVertexCount,a.maxIndexCount=this._maxIndexCount,a.geometryInitialized=this._geometryInitialized,a.geometryCount=this._geometryCount,a.matricesTexture=this._matricesTexture.toJSON(e),this._colorsTexture!==null&&(a.colorsTexture=this._colorsTexture.toJSON(e)),this.boundingSphere!==null&&(a.boundingSphere={center:a.boundingSphere.center.toArray(),radius:a.boundingSphere.radius}),this.boundingBox!==null&&(a.boundingBox={min:a.boundingBox.min.toArray(),max:a.boundingBox.max.toArray()}));function l(f,h){return f[h.uuid]===void 0&&(f[h.uuid]=h.toJSON(e)),h.uuid}if(this.isScene)this.background&&(this.background.isColor?a.background=this.background.toJSON():this.background.isTexture&&(a.background=this.background.toJSON(e).uuid)),this.environment&&this.environment.isTexture&&this.environment.isRenderTargetTexture!==!0&&(a.environment=this.environment.toJSON(e).uuid);else if(this.isMesh||this.isLine||this.isPoints){a.geometry=l(e.geometries,this.geometry);const f=this.geometry.parameters;if(f!==void 0&&f.shapes!==void 0){const h=f.shapes;if(Array.isArray(h))for(let p=0,m=h.length;p<m;p++){const _=h[p];l(e.shapes,_)}else l(e.shapes,h)}}if(this.isSkinnedMesh&&(a.bindMode=this.bindMode,a.bindMatrix=this.bindMatrix.toArray(),this.skeleton!==void 0&&(l(e.skeletons,this.skeleton),a.skeleton=this.skeleton.uuid)),this.material!==void 0)if(Array.isArray(this.material)){const f=[];for(let h=0,p=this.material.length;h<p;h++)f.push(l(e.materials,this.material[h]));a.material=f}else a.material=l(e.materials,this.material);if(this.children.length>0){a.children=[];for(let f=0;f<this.children.length;f++)a.children.push(this.children[f].toJSON(e).object)}if(this.animations.length>0){a.animations=[];for(let f=0;f<this.animations.length;f++){const h=this.animations[f];a.animations.push(l(e.animations,h))}}if(t){const f=u(e.geometries),h=u(e.materials),p=u(e.textures),m=u(e.images),_=u(e.shapes),x=u(e.skeletons),S=u(e.animations),T=u(e.nodes);f.length>0&&(s.geometries=f),h.length>0&&(s.materials=h),p.length>0&&(s.textures=p),m.length>0&&(s.images=m),_.length>0&&(s.shapes=_),x.length>0&&(s.skeletons=x),S.length>0&&(s.animations=S),T.length>0&&(s.nodes=T)}return s.object=a,s;function u(f){const h=[];for(const p in f){const m=f[p];delete m.metadata,h.push(m)}return h}}clone(e){return new this.constructor().copy(this,e)}copy(e,t=!0){if(this.name=e.name,this.up.copy(e.up),this.position.copy(e.position),this.rotation.order=e.rotation.order,this.quaternion.copy(e.quaternion),this.scale.copy(e.scale),this.matrix.copy(e.matrix),this.matrixWorld.copy(e.matrixWorld),this.matrixAutoUpdate=e.matrixAutoUpdate,this.matrixWorldAutoUpdate=e.matrixWorldAutoUpdate,this.matrixWorldNeedsUpdate=e.matrixWorldNeedsUpdate,this.layers.mask=e.layers.mask,this.visible=e.visible,this.castShadow=e.castShadow,this.receiveShadow=e.receiveShadow,this.frustumCulled=e.frustumCulled,this.renderOrder=e.renderOrder,this.animations=e.animations.slice(),this.userData=JSON.parse(JSON.stringify(e.userData)),t===!0)for(let s=0;s<e.children.length;s++){const a=e.children[s];this.add(a.clone())}return this}}vn.DEFAULT_UP=new Q(0,1,0);vn.DEFAULT_MATRIX_AUTO_UPDATE=!0;vn.DEFAULT_MATRIX_WORLD_AUTO_UPDATE=!0;const li=new Q,Fi=new Q,ef=new Q,Oi=new Q,Vs=new Q,Gs=new Q,Im=new Q,tf=new Q,nf=new Q,rf=new Q,sf=new Vt,of=new Vt,af=new Vt;class fi{constructor(e=new Q,t=new Q,s=new Q){this.a=e,this.b=t,this.c=s}static getNormal(e,t,s,a){a.subVectors(s,t),li.subVectors(e,t),a.cross(li);const l=a.lengthSq();return l>0?a.multiplyScalar(1/Math.sqrt(l)):a.set(0,0,0)}static getBarycoord(e,t,s,a,l){li.subVectors(a,t),Fi.subVectors(s,t),ef.subVectors(e,t);const u=li.dot(li),f=li.dot(Fi),h=li.dot(ef),p=Fi.dot(Fi),m=Fi.dot(ef),_=u*p-f*f;if(_===0)return l.set(0,0,0),null;const x=1/_,S=(p*h-f*m)*x,T=(u*m-f*h)*x;return l.set(1-S-T,T,S)}static containsPoint(e,t,s,a){return this.getBarycoord(e,t,s,a,Oi)===null?!1:Oi.x>=0&&Oi.y>=0&&Oi.x+Oi.y<=1}static getInterpolation(e,t,s,a,l,u,f,h){return this.getBarycoord(e,t,s,a,Oi)===null?(h.x=0,h.y=0,"z"in h&&(h.z=0),"w"in h&&(h.w=0),null):(h.setScalar(0),h.addScaledVector(l,Oi.x),h.addScaledVector(u,Oi.y),h.addScaledVector(f,Oi.z),h)}static getInterpolatedAttribute(e,t,s,a,l,u){return sf.setScalar(0),of.setScalar(0),af.setScalar(0),sf.fromBufferAttribute(e,t),of.fromBufferAttribute(e,s),af.fromBufferAttribute(e,a),u.setScalar(0),u.addScaledVector(sf,l.x),u.addScaledVector(of,l.y),u.addScaledVector(af,l.z),u}static isFrontFacing(e,t,s,a){return li.subVectors(s,t),Fi.subVectors(e,t),li.cross(Fi).dot(a)<0}set(e,t,s){return this.a.copy(e),this.b.copy(t),this.c.copy(s),this}setFromPointsAndIndices(e,t,s,a){return this.a.copy(e[t]),this.b.copy(e[s]),this.c.copy(e[a]),this}setFromAttributeAndIndices(e,t,s,a){return this.a.fromBufferAttribute(e,t),this.b.fromBufferAttribute(e,s),this.c.fromBufferAttribute(e,a),this}clone(){return new this.constructor().copy(this)}copy(e){return this.a.copy(e.a),this.b.copy(e.b),this.c.copy(e.c),this}getArea(){return li.subVectors(this.c,this.b),Fi.subVectors(this.a,this.b),li.cross(Fi).length()*.5}getMidpoint(e){return e.addVectors(this.a,this.b).add(this.c).multiplyScalar(1/3)}getNormal(e){return fi.getNormal(this.a,this.b,this.c,e)}getPlane(e){return e.setFromCoplanarPoints(this.a,this.b,this.c)}getBarycoord(e,t){return fi.getBarycoord(e,this.a,this.b,this.c,t)}getInterpolation(e,t,s,a,l){return fi.getInterpolation(e,this.a,this.b,this.c,t,s,a,l)}containsPoint(e){return fi.containsPoint(e,this.a,this.b,this.c)}isFrontFacing(e){return fi.isFrontFacing(this.a,this.b,this.c,e)}intersectsBox(e){return e.intersectsTriangle(this)}closestPointToPoint(e,t){const s=this.a,a=this.b,l=this.c;let u,f;Vs.subVectors(a,s),Gs.subVectors(l,s),tf.subVectors(e,s);const h=Vs.dot(tf),p=Gs.dot(tf);if(h<=0&&p<=0)return t.copy(s);nf.subVectors(e,a);const m=Vs.dot(nf),_=Gs.dot(nf);if(m>=0&&_<=m)return t.copy(a);const x=h*_-m*p;if(x<=0&&h>=0&&m<=0)return u=h/(h-m),t.copy(s).addScaledVector(Vs,u);rf.subVectors(e,l);const S=Vs.dot(rf),T=Gs.dot(rf);if(T>=0&&S<=T)return t.copy(l);const w=S*p-h*T;if(w<=0&&p>=0&&T<=0)return f=p/(p-T),t.copy(s).addScaledVector(Gs,f);const v=m*T-S*_;if(v<=0&&_-m>=0&&S-T>=0)return Im.subVectors(l,a),f=(_-m)/(_-m+(S-T)),t.copy(a).addScaledVector(Im,f);const y=1/(v+w+x);return u=w*y,f=x*y,t.copy(s).addScaledVector(Vs,u).addScaledVector(Gs,f)}equals(e){return e.a.equals(this.a)&&e.b.equals(this.b)&&e.c.equals(this.c)}}const Qg={aliceblue:15792383,antiquewhite:16444375,aqua:65535,aquamarine:8388564,azure:15794175,beige:16119260,bisque:16770244,black:0,blanchedalmond:16772045,blue:255,blueviolet:9055202,brown:10824234,burlywood:14596231,cadetblue:6266528,chartreuse:8388352,chocolate:13789470,coral:16744272,cornflowerblue:6591981,cornsilk:16775388,crimson:14423100,cyan:65535,darkblue:139,darkcyan:35723,darkgoldenrod:12092939,darkgray:11119017,darkgreen:25600,darkgrey:11119017,darkkhaki:12433259,darkmagenta:9109643,darkolivegreen:5597999,darkorange:16747520,darkorchid:10040012,darkred:9109504,darksalmon:15308410,darkseagreen:9419919,darkslateblue:4734347,darkslategray:3100495,darkslategrey:3100495,darkturquoise:52945,darkviolet:9699539,deeppink:16716947,deepskyblue:49151,dimgray:6908265,dimgrey:6908265,dodgerblue:2003199,firebrick:11674146,floralwhite:16775920,forestgreen:2263842,fuchsia:16711935,gainsboro:14474460,ghostwhite:16316671,gold:16766720,goldenrod:14329120,gray:8421504,green:32768,greenyellow:11403055,grey:8421504,honeydew:15794160,hotpink:16738740,indianred:13458524,indigo:4915330,ivory:16777200,khaki:15787660,lavender:15132410,lavenderblush:16773365,lawngreen:8190976,lemonchiffon:16775885,lightblue:11393254,lightcoral:15761536,lightcyan:14745599,lightgoldenrodyellow:16448210,lightgray:13882323,lightgreen:9498256,lightgrey:13882323,lightpink:16758465,lightsalmon:16752762,lightseagreen:2142890,lightskyblue:8900346,lightslategray:7833753,lightslategrey:7833753,lightsteelblue:11584734,lightyellow:16777184,lime:65280,limegreen:3329330,linen:16445670,magenta:16711935,maroon:8388608,mediumaquamarine:6737322,mediumblue:205,mediumorchid:12211667,mediumpurple:9662683,mediumseagreen:3978097,mediumslateblue:8087790,mediumspringgreen:64154,mediumturquoise:4772300,mediumvioletred:13047173,midnightblue:1644912,mintcream:16121850,mistyrose:16770273,moccasin:16770229,navajowhite:16768685,navy:128,oldlace:16643558,olive:8421376,olivedrab:7048739,orange:16753920,orangered:16729344,orchid:14315734,palegoldenrod:15657130,palegreen:10025880,paleturquoise:11529966,palevioletred:14381203,papayawhip:16773077,peachpuff:16767673,peru:13468991,pink:16761035,plum:14524637,powderblue:11591910,purple:8388736,rebeccapurple:6697881,red:16711680,rosybrown:12357519,royalblue:4286945,saddlebrown:9127187,salmon:16416882,sandybrown:16032864,seagreen:3050327,seashell:16774638,sienna:10506797,silver:12632256,skyblue:8900331,slateblue:6970061,slategray:7372944,slategrey:7372944,snow:16775930,springgreen:65407,steelblue:4620980,tan:13808780,teal:32896,thistle:14204888,tomato:16737095,turquoise:4251856,violet:15631086,wheat:16113331,white:16777215,whitesmoke:16119285,yellow:16776960,yellowgreen:10145074},vr={h:0,s:0,l:0},El={h:0,s:0,l:0};function lf(r,e,t){return t<0&&(t+=1),t>1&&(t-=1),t<1/6?r+(e-r)*6*t:t<1/2?e:t<2/3?r+(e-r)*6*(2/3-t):r}class vt{constructor(e,t,s){return this.isColor=!0,this.r=1,this.g=1,this.b=1,this.set(e,t,s)}set(e,t,s){if(t===void 0&&s===void 0){const a=e;a&&a.isColor?this.copy(a):typeof a=="number"?this.setHex(a):typeof a=="string"&&this.setStyle(a)}else this.setRGB(e,t,s);return this}setScalar(e){return this.r=e,this.g=e,this.b=e,this}setHex(e,t=ci){return e=Math.floor(e),this.r=(e>>16&255)/255,this.g=(e>>8&255)/255,this.b=(e&255)/255,Tt.toWorkingColorSpace(this,t),this}setRGB(e,t,s,a=Tt.workingColorSpace){return this.r=e,this.g=t,this.b=s,Tt.toWorkingColorSpace(this,a),this}setHSL(e,t,s,a=Tt.workingColorSpace){if(e=ny(e,1),t=Mn(t,0,1),s=Mn(s,0,1),t===0)this.r=this.g=this.b=s;else{const l=s<=.5?s*(1+t):s+t-s*t,u=2*s-l;this.r=lf(u,l,e+1/3),this.g=lf(u,l,e),this.b=lf(u,l,e-1/3)}return Tt.toWorkingColorSpace(this,a),this}setStyle(e,t=ci){function s(l){l!==void 0&&parseFloat(l)<1&&console.warn("THREE.Color: Alpha component of "+e+" will be ignored.")}let a;if(a=/^(\w+)\(([^\)]*)\)/.exec(e)){let l;const u=a[1],f=a[2];switch(u){case"rgb":case"rgba":if(l=/^\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*(?:,\s*(\d*\.?\d+)\s*)?$/.exec(f))return s(l[4]),this.setRGB(Math.min(255,parseInt(l[1],10))/255,Math.min(255,parseInt(l[2],10))/255,Math.min(255,parseInt(l[3],10))/255,t);if(l=/^\s*(\d+)\%\s*,\s*(\d+)\%\s*,\s*(\d+)\%\s*(?:,\s*(\d*\.?\d+)\s*)?$/.exec(f))return s(l[4]),this.setRGB(Math.min(100,parseInt(l[1],10))/100,Math.min(100,parseInt(l[2],10))/100,Math.min(100,parseInt(l[3],10))/100,t);break;case"hsl":case"hsla":if(l=/^\s*(\d*\.?\d+)\s*,\s*(\d*\.?\d+)\%\s*,\s*(\d*\.?\d+)\%\s*(?:,\s*(\d*\.?\d+)\s*)?$/.exec(f))return s(l[4]),this.setHSL(parseFloat(l[1])/360,parseFloat(l[2])/100,parseFloat(l[3])/100,t);break;default:console.warn("THREE.Color: Unknown color model "+e)}}else if(a=/^\#([A-Fa-f\d]+)$/.exec(e)){const l=a[1],u=l.length;if(u===3)return this.setRGB(parseInt(l.charAt(0),16)/15,parseInt(l.charAt(1),16)/15,parseInt(l.charAt(2),16)/15,t);if(u===6)return this.setHex(parseInt(l,16),t);console.warn("THREE.Color: Invalid hex color "+e)}else if(e&&e.length>0)return this.setColorName(e,t);return this}setColorName(e,t=ci){const s=Qg[e.toLowerCase()];return s!==void 0?this.setHex(s,t):console.warn("THREE.Color: Unknown color "+e),this}clone(){return new this.constructor(this.r,this.g,this.b)}copy(e){return this.r=e.r,this.g=e.g,this.b=e.b,this}copySRGBToLinear(e){return this.r=Js(e.r),this.g=Js(e.g),this.b=Js(e.b),this}copyLinearToSRGB(e){return this.r=ju(e.r),this.g=ju(e.g),this.b=ju(e.b),this}convertSRGBToLinear(){return this.copySRGBToLinear(this),this}convertLinearToSRGB(){return this.copyLinearToSRGB(this),this}getHex(e=ci){return Tt.fromWorkingColorSpace(_n.copy(this),e),Math.round(Mn(_n.r*255,0,255))*65536+Math.round(Mn(_n.g*255,0,255))*256+Math.round(Mn(_n.b*255,0,255))}getHexString(e=ci){return("000000"+this.getHex(e).toString(16)).slice(-6)}getHSL(e,t=Tt.workingColorSpace){Tt.fromWorkingColorSpace(_n.copy(this),t);const s=_n.r,a=_n.g,l=_n.b,u=Math.max(s,a,l),f=Math.min(s,a,l);let h,p;const m=(f+u)/2;if(f===u)h=0,p=0;else{const _=u-f;switch(p=m<=.5?_/(u+f):_/(2-u-f),u){case s:h=(a-l)/_+(a<l?6:0);break;case a:h=(l-s)/_+2;break;case l:h=(s-a)/_+4;break}h/=6}return e.h=h,e.s=p,e.l=m,e}getRGB(e,t=Tt.workingColorSpace){return Tt.fromWorkingColorSpace(_n.copy(this),t),e.r=_n.r,e.g=_n.g,e.b=_n.b,e}getStyle(e=ci){Tt.fromWorkingColorSpace(_n.copy(this),e);const t=_n.r,s=_n.g,a=_n.b;return e!==ci?`color(${e} ${t.toFixed(3)} ${s.toFixed(3)} ${a.toFixed(3)})`:`rgb(${Math.round(t*255)},${Math.round(s*255)},${Math.round(a*255)})`}offsetHSL(e,t,s){return this.getHSL(vr),this.setHSL(vr.h+e,vr.s+t,vr.l+s)}add(e){return this.r+=e.r,this.g+=e.g,this.b+=e.b,this}addColors(e,t){return this.r=e.r+t.r,this.g=e.g+t.g,this.b=e.b+t.b,this}addScalar(e){return this.r+=e,this.g+=e,this.b+=e,this}sub(e){return this.r=Math.max(0,this.r-e.r),this.g=Math.max(0,this.g-e.g),this.b=Math.max(0,this.b-e.b),this}multiply(e){return this.r*=e.r,this.g*=e.g,this.b*=e.b,this}multiplyScalar(e){return this.r*=e,this.g*=e,this.b*=e,this}lerp(e,t){return this.r+=(e.r-this.r)*t,this.g+=(e.g-this.g)*t,this.b+=(e.b-this.b)*t,this}lerpColors(e,t,s){return this.r=e.r+(t.r-e.r)*s,this.g=e.g+(t.g-e.g)*s,this.b=e.b+(t.b-e.b)*s,this}lerpHSL(e,t){this.getHSL(vr),e.getHSL(El);const s=Gu(vr.h,El.h,t),a=Gu(vr.s,El.s,t),l=Gu(vr.l,El.l,t);return this.setHSL(s,a,l),this}setFromVector3(e){return this.r=e.x,this.g=e.y,this.b=e.z,this}applyMatrix3(e){const t=this.r,s=this.g,a=this.b,l=e.elements;return this.r=l[0]*t+l[3]*s+l[6]*a,this.g=l[1]*t+l[4]*s+l[7]*a,this.b=l[2]*t+l[5]*s+l[8]*a,this}equals(e){return e.r===this.r&&e.g===this.g&&e.b===this.b}fromArray(e,t=0){return this.r=e[t],this.g=e[t+1],this.b=e[t+2],this}toArray(e=[],t=0){return e[t]=this.r,e[t+1]=this.g,e[t+2]=this.b,e}fromBufferAttribute(e,t){return this.r=e.getX(t),this.g=e.getY(t),this.b=e.getZ(t),this}toJSON(){return this.getHex()}*[Symbol.iterator](){yield this.r,yield this.g,yield this.b}}const _n=new vt;vt.NAMES=Qg;let Sy=0;class aa extends os{constructor(){super(),this.isMaterial=!0,Object.defineProperty(this,"id",{value:Sy++}),this.uuid=oa(),this.name="",this.type="Material",this.blending=Zs,this.side=Ar,this.vertexColors=!1,this.opacity=1,this.transparent=!1,this.alphaHash=!1,this.blendSrc=wf,this.blendDst=Af,this.blendEquation=Qr,this.blendSrcAlpha=null,this.blendDstAlpha=null,this.blendEquationAlpha=null,this.blendColor=new vt(0,0,0),this.blendAlpha=0,this.depthFunc=eo,this.depthTest=!0,this.depthWrite=!0,this.stencilWriteMask=255,this.stencilFunc=ym,this.stencilRef=0,this.stencilFuncMask=255,this.stencilFail=Is,this.stencilZFail=Is,this.stencilZPass=Is,this.stencilWrite=!1,this.clippingPlanes=null,this.clipIntersection=!1,this.clipShadows=!1,this.shadowSide=null,this.colorWrite=!0,this.precision=null,this.polygonOffset=!1,this.polygonOffsetFactor=0,this.polygonOffsetUnits=0,this.dithering=!1,this.alphaToCoverage=!1,this.premultipliedAlpha=!1,this.forceSinglePass=!1,this.visible=!0,this.toneMapped=!0,this.userData={},this.version=0,this._alphaTest=0}get alphaTest(){return this._alphaTest}set alphaTest(e){this._alphaTest>0!=e>0&&this.version++,this._alphaTest=e}onBeforeRender(){}onBeforeCompile(){}customProgramCacheKey(){return this.onBeforeCompile.toString()}setValues(e){if(e!==void 0)for(const t in e){const s=e[t];if(s===void 0){console.warn(`THREE.Material: parameter '${t}' has value of undefined.`);continue}const a=this[t];if(a===void 0){console.warn(`THREE.Material: '${t}' is not a property of THREE.${this.type}.`);continue}a&&a.isColor?a.set(s):a&&a.isVector3&&s&&s.isVector3?a.copy(s):this[t]=s}}toJSON(e){const t=e===void 0||typeof e=="string";t&&(e={textures:{},images:{}});const s={metadata:{version:4.6,type:"Material",generator:"Material.toJSON"}};s.uuid=this.uuid,s.type=this.type,this.name!==""&&(s.name=this.name),this.color&&this.color.isColor&&(s.color=this.color.getHex()),this.roughness!==void 0&&(s.roughness=this.roughness),this.metalness!==void 0&&(s.metalness=this.metalness),this.sheen!==void 0&&(s.sheen=this.sheen),this.sheenColor&&this.sheenColor.isColor&&(s.sheenColor=this.sheenColor.getHex()),this.sheenRoughness!==void 0&&(s.sheenRoughness=this.sheenRoughness),this.emissive&&this.emissive.isColor&&(s.emissive=this.emissive.getHex()),this.emissiveIntensity!==void 0&&this.emissiveIntensity!==1&&(s.emissiveIntensity=this.emissiveIntensity),this.specular&&this.specular.isColor&&(s.specular=this.specular.getHex()),this.specularIntensity!==void 0&&(s.specularIntensity=this.specularIntensity),this.specularColor&&this.specularColor.isColor&&(s.specularColor=this.specularColor.getHex()),this.shininess!==void 0&&(s.shininess=this.shininess),this.clearcoat!==void 0&&(s.clearcoat=this.clearcoat),this.clearcoatRoughness!==void 0&&(s.clearcoatRoughness=this.clearcoatRoughness),this.clearcoatMap&&this.clearcoatMap.isTexture&&(s.clearcoatMap=this.clearcoatMap.toJSON(e).uuid),this.clearcoatRoughnessMap&&this.clearcoatRoughnessMap.isTexture&&(s.clearcoatRoughnessMap=this.clearcoatRoughnessMap.toJSON(e).uuid),this.clearcoatNormalMap&&this.clearcoatNormalMap.isTexture&&(s.clearcoatNormalMap=this.clearcoatNormalMap.toJSON(e).uuid,s.clearcoatNormalScale=this.clearcoatNormalScale.toArray()),this.dispersion!==void 0&&(s.dispersion=this.dispersion),this.iridescence!==void 0&&(s.iridescence=this.iridescence),this.iridescenceIOR!==void 0&&(s.iridescenceIOR=this.iridescenceIOR),this.iridescenceThicknessRange!==void 0&&(s.iridescenceThicknessRange=this.iridescenceThicknessRange),this.iridescenceMap&&this.iridescenceMap.isTexture&&(s.iridescenceMap=this.iridescenceMap.toJSON(e).uuid),this.iridescenceThicknessMap&&this.iridescenceThicknessMap.isTexture&&(s.iridescenceThicknessMap=this.iridescenceThicknessMap.toJSON(e).uuid),this.anisotropy!==void 0&&(s.anisotropy=this.anisotropy),this.anisotropyRotation!==void 0&&(s.anisotropyRotation=this.anisotropyRotation),this.anisotropyMap&&this.anisotropyMap.isTexture&&(s.anisotropyMap=this.anisotropyMap.toJSON(e).uuid),this.map&&this.map.isTexture&&(s.map=this.map.toJSON(e).uuid),this.matcap&&this.matcap.isTexture&&(s.matcap=this.matcap.toJSON(e).uuid),this.alphaMap&&this.alphaMap.isTexture&&(s.alphaMap=this.alphaMap.toJSON(e).uuid),this.lightMap&&this.lightMap.isTexture&&(s.lightMap=this.lightMap.toJSON(e).uuid,s.lightMapIntensity=this.lightMapIntensity),this.aoMap&&this.aoMap.isTexture&&(s.aoMap=this.aoMap.toJSON(e).uuid,s.aoMapIntensity=this.aoMapIntensity),this.bumpMap&&this.bumpMap.isTexture&&(s.bumpMap=this.bumpMap.toJSON(e).uuid,s.bumpScale=this.bumpScale),this.normalMap&&this.normalMap.isTexture&&(s.normalMap=this.normalMap.toJSON(e).uuid,s.normalMapType=this.normalMapType,s.normalScale=this.normalScale.toArray()),this.displacementMap&&this.displacementMap.isTexture&&(s.displacementMap=this.displacementMap.toJSON(e).uuid,s.displacementScale=this.displacementScale,s.displacementBias=this.displacementBias),this.roughnessMap&&this.roughnessMap.isTexture&&(s.roughnessMap=this.roughnessMap.toJSON(e).uuid),this.metalnessMap&&this.metalnessMap.isTexture&&(s.metalnessMap=this.metalnessMap.toJSON(e).uuid),this.emissiveMap&&this.emissiveMap.isTexture&&(s.emissiveMap=this.emissiveMap.toJSON(e).uuid),this.specularMap&&this.specularMap.isTexture&&(s.specularMap=this.specularMap.toJSON(e).uuid),this.specularIntensityMap&&this.specularIntensityMap.isTexture&&(s.specularIntensityMap=this.specularIntensityMap.toJSON(e).uuid),this.specularColorMap&&this.specularColorMap.isTexture&&(s.specularColorMap=this.specularColorMap.toJSON(e).uuid),this.envMap&&this.envMap.isTexture&&(s.envMap=this.envMap.toJSON(e).uuid,this.combine!==void 0&&(s.combine=this.combine)),this.envMapRotation!==void 0&&(s.envMapRotation=this.envMapRotation.toArray()),this.envMapIntensity!==void 0&&(s.envMapIntensity=this.envMapIntensity),this.reflectivity!==void 0&&(s.reflectivity=this.reflectivity),this.refractionRatio!==void 0&&(s.refractionRatio=this.refractionRatio),this.gradientMap&&this.gradientMap.isTexture&&(s.gradientMap=this.gradientMap.toJSON(e).uuid),this.transmission!==void 0&&(s.transmission=this.transmission),this.transmissionMap&&this.transmissionMap.isTexture&&(s.transmissionMap=this.transmissionMap.toJSON(e).uuid),this.thickness!==void 0&&(s.thickness=this.thickness),this.thicknessMap&&this.thicknessMap.isTexture&&(s.thicknessMap=this.thicknessMap.toJSON(e).uuid),this.attenuationDistance!==void 0&&this.attenuationDistance!==1/0&&(s.attenuationDistance=this.attenuationDistance),this.attenuationColor!==void 0&&(s.attenuationColor=this.attenuationColor.getHex()),this.size!==void 0&&(s.size=this.size),this.shadowSide!==null&&(s.shadowSide=this.shadowSide),this.sizeAttenuation!==void 0&&(s.sizeAttenuation=this.sizeAttenuation),this.blending!==Zs&&(s.blending=this.blending),this.side!==Ar&&(s.side=this.side),this.vertexColors===!0&&(s.vertexColors=!0),this.opacity<1&&(s.opacity=this.opacity),this.transparent===!0&&(s.transparent=!0),this.blendSrc!==wf&&(s.blendSrc=this.blendSrc),this.blendDst!==Af&&(s.blendDst=this.blendDst),this.blendEquation!==Qr&&(s.blendEquation=this.blendEquation),this.blendSrcAlpha!==null&&(s.blendSrcAlpha=this.blendSrcAlpha),this.blendDstAlpha!==null&&(s.blendDstAlpha=this.blendDstAlpha),this.blendEquationAlpha!==null&&(s.blendEquationAlpha=this.blendEquationAlpha),this.blendColor&&this.blendColor.isColor&&(s.blendColor=this.blendColor.getHex()),this.blendAlpha!==0&&(s.blendAlpha=this.blendAlpha),this.depthFunc!==eo&&(s.depthFunc=this.depthFunc),this.depthTest===!1&&(s.depthTest=this.depthTest),this.depthWrite===!1&&(s.depthWrite=this.depthWrite),this.colorWrite===!1&&(s.colorWrite=this.colorWrite),this.stencilWriteMask!==255&&(s.stencilWriteMask=this.stencilWriteMask),this.stencilFunc!==ym&&(s.stencilFunc=this.stencilFunc),this.stencilRef!==0&&(s.stencilRef=this.stencilRef),this.stencilFuncMask!==255&&(s.stencilFuncMask=this.stencilFuncMask),this.stencilFail!==Is&&(s.stencilFail=this.stencilFail),this.stencilZFail!==Is&&(s.stencilZFail=this.stencilZFail),this.stencilZPass!==Is&&(s.stencilZPass=this.stencilZPass),this.stencilWrite===!0&&(s.stencilWrite=this.stencilWrite),this.rotation!==void 0&&this.rotation!==0&&(s.rotation=this.rotation),this.polygonOffset===!0&&(s.polygonOffset=!0),this.polygonOffsetFactor!==0&&(s.polygonOffsetFactor=this.polygonOffsetFactor),this.polygonOffsetUnits!==0&&(s.polygonOffsetUnits=this.polygonOffsetUnits),this.linewidth!==void 0&&this.linewidth!==1&&(s.linewidth=this.linewidth),this.dashSize!==void 0&&(s.dashSize=this.dashSize),this.gapSize!==void 0&&(s.gapSize=this.gapSize),this.scale!==void 0&&(s.scale=this.scale),this.dithering===!0&&(s.dithering=!0),this.alphaTest>0&&(s.alphaTest=this.alphaTest),this.alphaHash===!0&&(s.alphaHash=!0),this.alphaToCoverage===!0&&(s.alphaToCoverage=!0),this.premultipliedAlpha===!0&&(s.premultipliedAlpha=!0),this.forceSinglePass===!0&&(s.forceSinglePass=!0),this.wireframe===!0&&(s.wireframe=!0),this.wireframeLinewidth>1&&(s.wireframeLinewidth=this.wireframeLinewidth),this.wireframeLinecap!=="round"&&(s.wireframeLinecap=this.wireframeLinecap),this.wireframeLinejoin!=="round"&&(s.wireframeLinejoin=this.wireframeLinejoin),this.flatShading===!0&&(s.flatShading=!0),this.visible===!1&&(s.visible=!1),this.toneMapped===!1&&(s.toneMapped=!1),this.fog===!1&&(s.fog=!1),Object.keys(this.userData).length>0&&(s.userData=this.userData);function a(l){const u=[];for(const f in l){const h=l[f];delete h.metadata,u.push(h)}return u}if(t){const l=a(e.textures),u=a(e.images);l.length>0&&(s.textures=l),u.length>0&&(s.images=u)}return s}clone(){return new this.constructor().copy(this)}copy(e){this.name=e.name,this.blending=e.blending,this.side=e.side,this.vertexColors=e.vertexColors,this.opacity=e.opacity,this.transparent=e.transparent,this.blendSrc=e.blendSrc,this.blendDst=e.blendDst,this.blendEquation=e.blendEquation,this.blendSrcAlpha=e.blendSrcAlpha,this.blendDstAlpha=e.blendDstAlpha,this.blendEquationAlpha=e.blendEquationAlpha,this.blendColor.copy(e.blendColor),this.blendAlpha=e.blendAlpha,this.depthFunc=e.depthFunc,this.depthTest=e.depthTest,this.depthWrite=e.depthWrite,this.stencilWriteMask=e.stencilWriteMask,this.stencilFunc=e.stencilFunc,this.stencilRef=e.stencilRef,this.stencilFuncMask=e.stencilFuncMask,this.stencilFail=e.stencilFail,this.stencilZFail=e.stencilZFail,this.stencilZPass=e.stencilZPass,this.stencilWrite=e.stencilWrite;const t=e.clippingPlanes;let s=null;if(t!==null){const a=t.length;s=new Array(a);for(let l=0;l!==a;++l)s[l]=t[l].clone()}return this.clippingPlanes=s,this.clipIntersection=e.clipIntersection,this.clipShadows=e.clipShadows,this.shadowSide=e.shadowSide,this.colorWrite=e.colorWrite,this.precision=e.precision,this.polygonOffset=e.polygonOffset,this.polygonOffsetFactor=e.polygonOffsetFactor,this.polygonOffsetUnits=e.polygonOffsetUnits,this.dithering=e.dithering,this.alphaTest=e.alphaTest,this.alphaHash=e.alphaHash,this.alphaToCoverage=e.alphaToCoverage,this.premultipliedAlpha=e.premultipliedAlpha,this.forceSinglePass=e.forceSinglePass,this.visible=e.visible,this.toneMapped=e.toneMapped,this.userData=JSON.parse(JSON.stringify(e.userData)),this}dispose(){this.dispatchEvent({type:"dispose"})}set needsUpdate(e){e===!0&&this.version++}onBuild(){console.warn("Material: onBuild() has been removed.")}}class Jg extends aa{constructor(e){super(),this.isMeshBasicMaterial=!0,this.type="MeshBasicMaterial",this.color=new vt(16777215),this.map=null,this.lightMap=null,this.lightMapIntensity=1,this.aoMap=null,this.aoMapIntensity=1,this.specularMap=null,this.alphaMap=null,this.envMap=null,this.envMapRotation=new Si,this.combine=Ng,this.reflectivity=1,this.refractionRatio=.98,this.wireframe=!1,this.wireframeLinewidth=1,this.wireframeLinecap="round",this.wireframeLinejoin="round",this.fog=!0,this.setValues(e)}copy(e){return super.copy(e),this.color.copy(e.color),this.map=e.map,this.lightMap=e.lightMap,this.lightMapIntensity=e.lightMapIntensity,this.aoMap=e.aoMap,this.aoMapIntensity=e.aoMapIntensity,this.specularMap=e.specularMap,this.alphaMap=e.alphaMap,this.envMap=e.envMap,this.envMapRotation.copy(e.envMapRotation),this.combine=e.combine,this.reflectivity=e.reflectivity,this.refractionRatio=e.refractionRatio,this.wireframe=e.wireframe,this.wireframeLinewidth=e.wireframeLinewidth,this.wireframeLinecap=e.wireframeLinecap,this.wireframeLinejoin=e.wireframeLinejoin,this.fog=e.fog,this}}const Xt=new Q,Tl=new rt;class Vn{constructor(e,t,s=!1){if(Array.isArray(e))throw new TypeError("THREE.BufferAttribute: array should be a Typed Array.");this.isBufferAttribute=!0,this.name="",this.array=e,this.itemSize=t,this.count=e!==void 0?e.length/t:0,this.normalized=s,this.usage=Sm,this.updateRanges=[],this.gpuType=Hi,this.version=0}onUploadCallback(){}set needsUpdate(e){e===!0&&this.version++}setUsage(e){return this.usage=e,this}addUpdateRange(e,t){this.updateRanges.push({start:e,count:t})}clearUpdateRanges(){this.updateRanges.length=0}copy(e){return this.name=e.name,this.array=new e.array.constructor(e.array),this.itemSize=e.itemSize,this.count=e.count,this.normalized=e.normalized,this.usage=e.usage,this.gpuType=e.gpuType,this}copyAt(e,t,s){e*=this.itemSize,s*=t.itemSize;for(let a=0,l=this.itemSize;a<l;a++)this.array[e+a]=t.array[s+a];return this}copyArray(e){return this.array.set(e),this}applyMatrix3(e){if(this.itemSize===2)for(let t=0,s=this.count;t<s;t++)Tl.fromBufferAttribute(this,t),Tl.applyMatrix3(e),this.setXY(t,Tl.x,Tl.y);else if(this.itemSize===3)for(let t=0,s=this.count;t<s;t++)Xt.fromBufferAttribute(this,t),Xt.applyMatrix3(e),this.setXYZ(t,Xt.x,Xt.y,Xt.z);return this}applyMatrix4(e){for(let t=0,s=this.count;t<s;t++)Xt.fromBufferAttribute(this,t),Xt.applyMatrix4(e),this.setXYZ(t,Xt.x,Xt.y,Xt.z);return this}applyNormalMatrix(e){for(let t=0,s=this.count;t<s;t++)Xt.fromBufferAttribute(this,t),Xt.applyNormalMatrix(e),this.setXYZ(t,Xt.x,Xt.y,Xt.z);return this}transformDirection(e){for(let t=0,s=this.count;t<s;t++)Xt.fromBufferAttribute(this,t),Xt.transformDirection(e),this.setXYZ(t,Xt.x,Xt.y,Xt.z);return this}set(e,t=0){return this.array.set(e,t),this}getComponent(e,t){let s=this.array[e*this.itemSize+t];return this.normalized&&(s=Xo(s,this.array)),s}setComponent(e,t,s){return this.normalized&&(s=bn(s,this.array)),this.array[e*this.itemSize+t]=s,this}getX(e){let t=this.array[e*this.itemSize];return this.normalized&&(t=Xo(t,this.array)),t}setX(e,t){return this.normalized&&(t=bn(t,this.array)),this.array[e*this.itemSize]=t,this}getY(e){let t=this.array[e*this.itemSize+1];return this.normalized&&(t=Xo(t,this.array)),t}setY(e,t){return this.normalized&&(t=bn(t,this.array)),this.array[e*this.itemSize+1]=t,this}getZ(e){let t=this.array[e*this.itemSize+2];return this.normalized&&(t=Xo(t,this.array)),t}setZ(e,t){return this.normalized&&(t=bn(t,this.array)),this.array[e*this.itemSize+2]=t,this}getW(e){let t=this.array[e*this.itemSize+3];return this.normalized&&(t=Xo(t,this.array)),t}setW(e,t){return this.normalized&&(t=bn(t,this.array)),this.array[e*this.itemSize+3]=t,this}setXY(e,t,s){return e*=this.itemSize,this.normalized&&(t=bn(t,this.array),s=bn(s,this.array)),this.array[e+0]=t,this.array[e+1]=s,this}setXYZ(e,t,s,a){return e*=this.itemSize,this.normalized&&(t=bn(t,this.array),s=bn(s,this.array),a=bn(a,this.array)),this.array[e+0]=t,this.array[e+1]=s,this.array[e+2]=a,this}setXYZW(e,t,s,a,l){return e*=this.itemSize,this.normalized&&(t=bn(t,this.array),s=bn(s,this.array),a=bn(a,this.array),l=bn(l,this.array)),this.array[e+0]=t,this.array[e+1]=s,this.array[e+2]=a,this.array[e+3]=l,this}onUpload(e){return this.onUploadCallback=e,this}clone(){return new this.constructor(this.array,this.itemSize).copy(this)}toJSON(){const e={itemSize:this.itemSize,type:this.array.constructor.name,array:Array.from(this.array),normalized:this.normalized};return this.name!==""&&(e.name=this.name),this.usage!==Sm&&(e.usage=this.usage),e}}class e_ extends Vn{constructor(e,t,s){super(new Uint16Array(e),t,s)}}class t_ extends Vn{constructor(e,t,s){super(new Uint32Array(e),t,s)}}class Gi extends Vn{constructor(e,t,s){super(new Float32Array(e),t,s)}}let My=0;const Kn=new Gt,cf=new vn,Ws=new Q,Hn=new ao,Zo=new ao,on=new Q;class ji extends os{constructor(){super(),this.isBufferGeometry=!0,Object.defineProperty(this,"id",{value:My++}),this.uuid=oa(),this.name="",this.type="BufferGeometry",this.index=null,this.attributes={},this.morphAttributes={},this.morphTargetsRelative=!1,this.groups=[],this.boundingBox=null,this.boundingSphere=null,this.drawRange={start:0,count:1/0},this.userData={}}getIndex(){return this.index}setIndex(e){return Array.isArray(e)?this.index=new(Yg(e)?t_:e_)(e,1):this.index=e,this}getAttribute(e){return this.attributes[e]}setAttribute(e,t){return this.attributes[e]=t,this}deleteAttribute(e){return delete this.attributes[e],this}hasAttribute(e){return this.attributes[e]!==void 0}addGroup(e,t,s=0){this.groups.push({start:e,count:t,materialIndex:s})}clearGroups(){this.groups=[]}setDrawRange(e,t){this.drawRange.start=e,this.drawRange.count=t}applyMatrix4(e){const t=this.attributes.position;t!==void 0&&(t.applyMatrix4(e),t.needsUpdate=!0);const s=this.attributes.normal;if(s!==void 0){const l=new ot().getNormalMatrix(e);s.applyNormalMatrix(l),s.needsUpdate=!0}const a=this.attributes.tangent;return a!==void 0&&(a.transformDirection(e),a.needsUpdate=!0),this.boundingBox!==null&&this.computeBoundingBox(),this.boundingSphere!==null&&this.computeBoundingSphere(),this}applyQuaternion(e){return Kn.makeRotationFromQuaternion(e),this.applyMatrix4(Kn),this}rotateX(e){return Kn.makeRotationX(e),this.applyMatrix4(Kn),this}rotateY(e){return Kn.makeRotationY(e),this.applyMatrix4(Kn),this}rotateZ(e){return Kn.makeRotationZ(e),this.applyMatrix4(Kn),this}translate(e,t,s){return Kn.makeTranslation(e,t,s),this.applyMatrix4(Kn),this}scale(e,t,s){return Kn.makeScale(e,t,s),this.applyMatrix4(Kn),this}lookAt(e){return cf.lookAt(e),cf.updateMatrix(),this.applyMatrix4(cf.matrix),this}center(){return this.computeBoundingBox(),this.boundingBox.getCenter(Ws).negate(),this.translate(Ws.x,Ws.y,Ws.z),this}setFromPoints(e){const t=[];for(let s=0,a=e.length;s<a;s++){const l=e[s];t.push(l.x,l.y,l.z||0)}return this.setAttribute("position",new Gi(t,3)),this}computeBoundingBox(){this.boundingBox===null&&(this.boundingBox=new ao);const e=this.attributes.position,t=this.morphAttributes.position;if(e&&e.isGLBufferAttribute){console.error("THREE.BufferGeometry.computeBoundingBox(): GLBufferAttribute requires a manual bounding box.",this),this.boundingBox.set(new Q(-1/0,-1/0,-1/0),new Q(1/0,1/0,1/0));return}if(e!==void 0){if(this.boundingBox.setFromBufferAttribute(e),t)for(let s=0,a=t.length;s<a;s++){const l=t[s];Hn.setFromBufferAttribute(l),this.morphTargetsRelative?(on.addVectors(this.boundingBox.min,Hn.min),this.boundingBox.expandByPoint(on),on.addVectors(this.boundingBox.max,Hn.max),this.boundingBox.expandByPoint(on)):(this.boundingBox.expandByPoint(Hn.min),this.boundingBox.expandByPoint(Hn.max))}}else this.boundingBox.makeEmpty();(isNaN(this.boundingBox.min.x)||isNaN(this.boundingBox.min.y)||isNaN(this.boundingBox.min.z))&&console.error('THREE.BufferGeometry.computeBoundingBox(): Computed min/max have NaN values. The "position" attribute is likely to have NaN values.',this)}computeBoundingSphere(){this.boundingSphere===null&&(this.boundingSphere=new Td);const e=this.attributes.position,t=this.morphAttributes.position;if(e&&e.isGLBufferAttribute){console.error("THREE.BufferGeometry.computeBoundingSphere(): GLBufferAttribute requires a manual bounding sphere.",this),this.boundingSphere.set(new Q,1/0);return}if(e){const s=this.boundingSphere.center;if(Hn.setFromBufferAttribute(e),t)for(let l=0,u=t.length;l<u;l++){const f=t[l];Zo.setFromBufferAttribute(f),this.morphTargetsRelative?(on.addVectors(Hn.min,Zo.min),Hn.expandByPoint(on),on.addVectors(Hn.max,Zo.max),Hn.expandByPoint(on)):(Hn.expandByPoint(Zo.min),Hn.expandByPoint(Zo.max))}Hn.getCenter(s);let a=0;for(let l=0,u=e.count;l<u;l++)on.fromBufferAttribute(e,l),a=Math.max(a,s.distanceToSquared(on));if(t)for(let l=0,u=t.length;l<u;l++){const f=t[l],h=this.morphTargetsRelative;for(let p=0,m=f.count;p<m;p++)on.fromBufferAttribute(f,p),h&&(Ws.fromBufferAttribute(e,p),on.add(Ws)),a=Math.max(a,s.distanceToSquared(on))}this.boundingSphere.radius=Math.sqrt(a),isNaN(this.boundingSphere.radius)&&console.error('THREE.BufferGeometry.computeBoundingSphere(): Computed radius is NaN. The "position" attribute is likely to have NaN values.',this)}}computeTangents(){const e=this.index,t=this.attributes;if(e===null||t.position===void 0||t.normal===void 0||t.uv===void 0){console.error("THREE.BufferGeometry: .computeTangents() failed. Missing required attributes (index, position, normal or uv)");return}const s=t.position,a=t.normal,l=t.uv;this.hasAttribute("tangent")===!1&&this.setAttribute("tangent",new Vn(new Float32Array(4*s.count),4));const u=this.getAttribute("tangent"),f=[],h=[];for(let Y=0;Y<s.count;Y++)f[Y]=new Q,h[Y]=new Q;const p=new Q,m=new Q,_=new Q,x=new rt,S=new rt,T=new rt,w=new Q,v=new Q;function y(Y,ce,E){p.fromBufferAttribute(s,Y),m.fromBufferAttribute(s,ce),_.fromBufferAttribute(s,E),x.fromBufferAttribute(l,Y),S.fromBufferAttribute(l,ce),T.fromBufferAttribute(l,E),m.sub(p),_.sub(p),S.sub(x),T.sub(x);const C=1/(S.x*T.y-T.x*S.y);isFinite(C)&&(w.copy(m).multiplyScalar(T.y).addScaledVector(_,-S.y).multiplyScalar(C),v.copy(_).multiplyScalar(S.x).addScaledVector(m,-T.x).multiplyScalar(C),f[Y].add(w),f[ce].add(w),f[E].add(w),h[Y].add(v),h[ce].add(v),h[E].add(v))}let P=this.groups;P.length===0&&(P=[{start:0,count:e.count}]);for(let Y=0,ce=P.length;Y<ce;++Y){const E=P[Y],C=E.start,re=E.count;for(let ee=C,ae=C+re;ee<ae;ee+=3)y(e.getX(ee+0),e.getX(ee+1),e.getX(ee+2))}const b=new Q,D=new Q,V=new Q,O=new Q;function U(Y){V.fromBufferAttribute(a,Y),O.copy(V);const ce=f[Y];b.copy(ce),b.sub(V.multiplyScalar(V.dot(ce))).normalize(),D.crossVectors(O,ce);const C=D.dot(h[Y])<0?-1:1;u.setXYZW(Y,b.x,b.y,b.z,C)}for(let Y=0,ce=P.length;Y<ce;++Y){const E=P[Y],C=E.start,re=E.count;for(let ee=C,ae=C+re;ee<ae;ee+=3)U(e.getX(ee+0)),U(e.getX(ee+1)),U(e.getX(ee+2))}}computeVertexNormals(){const e=this.index,t=this.getAttribute("position");if(t!==void 0){let s=this.getAttribute("normal");if(s===void 0)s=new Vn(new Float32Array(t.count*3),3),this.setAttribute("normal",s);else for(let x=0,S=s.count;x<S;x++)s.setXYZ(x,0,0,0);const a=new Q,l=new Q,u=new Q,f=new Q,h=new Q,p=new Q,m=new Q,_=new Q;if(e)for(let x=0,S=e.count;x<S;x+=3){const T=e.getX(x+0),w=e.getX(x+1),v=e.getX(x+2);a.fromBufferAttribute(t,T),l.fromBufferAttribute(t,w),u.fromBufferAttribute(t,v),m.subVectors(u,l),_.subVectors(a,l),m.cross(_),f.fromBufferAttribute(s,T),h.fromBufferAttribute(s,w),p.fromBufferAttribute(s,v),f.add(m),h.add(m),p.add(m),s.setXYZ(T,f.x,f.y,f.z),s.setXYZ(w,h.x,h.y,h.z),s.setXYZ(v,p.x,p.y,p.z)}else for(let x=0,S=t.count;x<S;x+=3)a.fromBufferAttribute(t,x+0),l.fromBufferAttribute(t,x+1),u.fromBufferAttribute(t,x+2),m.subVectors(u,l),_.subVectors(a,l),m.cross(_),s.setXYZ(x+0,m.x,m.y,m.z),s.setXYZ(x+1,m.x,m.y,m.z),s.setXYZ(x+2,m.x,m.y,m.z);this.normalizeNormals(),s.needsUpdate=!0}}normalizeNormals(){const e=this.attributes.normal;for(let t=0,s=e.count;t<s;t++)on.fromBufferAttribute(e,t),on.normalize(),e.setXYZ(t,on.x,on.y,on.z)}toNonIndexed(){function e(f,h){const p=f.array,m=f.itemSize,_=f.normalized,x=new p.constructor(h.length*m);let S=0,T=0;for(let w=0,v=h.length;w<v;w++){f.isInterleavedBufferAttribute?S=h[w]*f.data.stride+f.offset:S=h[w]*m;for(let y=0;y<m;y++)x[T++]=p[S++]}return new Vn(x,m,_)}if(this.index===null)return console.warn("THREE.BufferGeometry.toNonIndexed(): BufferGeometry is already non-indexed."),this;const t=new ji,s=this.index.array,a=this.attributes;for(const f in a){const h=a[f],p=e(h,s);t.setAttribute(f,p)}const l=this.morphAttributes;for(const f in l){const h=[],p=l[f];for(let m=0,_=p.length;m<_;m++){const x=p[m],S=e(x,s);h.push(S)}t.morphAttributes[f]=h}t.morphTargetsRelative=this.morphTargetsRelative;const u=this.groups;for(let f=0,h=u.length;f<h;f++){const p=u[f];t.addGroup(p.start,p.count,p.materialIndex)}return t}toJSON(){const e={metadata:{version:4.6,type:"BufferGeometry",generator:"BufferGeometry.toJSON"}};if(e.uuid=this.uuid,e.type=this.type,this.name!==""&&(e.name=this.name),Object.keys(this.userData).length>0&&(e.userData=this.userData),this.parameters!==void 0){const h=this.parameters;for(const p in h)h[p]!==void 0&&(e[p]=h[p]);return e}e.data={attributes:{}};const t=this.index;t!==null&&(e.data.index={type:t.array.constructor.name,array:Array.prototype.slice.call(t.array)});const s=this.attributes;for(const h in s){const p=s[h];e.data.attributes[h]=p.toJSON(e.data)}const a={};let l=!1;for(const h in this.morphAttributes){const p=this.morphAttributes[h],m=[];for(let _=0,x=p.length;_<x;_++){const S=p[_];m.push(S.toJSON(e.data))}m.length>0&&(a[h]=m,l=!0)}l&&(e.data.morphAttributes=a,e.data.morphTargetsRelative=this.morphTargetsRelative);const u=this.groups;u.length>0&&(e.data.groups=JSON.parse(JSON.stringify(u)));const f=this.boundingSphere;return f!==null&&(e.data.boundingSphere={center:f.center.toArray(),radius:f.radius}),e}clone(){return new this.constructor().copy(this)}copy(e){this.index=null,this.attributes={},this.morphAttributes={},this.groups=[],this.boundingBox=null,this.boundingSphere=null;const t={};this.name=e.name;const s=e.index;s!==null&&this.setIndex(s.clone(t));const a=e.attributes;for(const p in a){const m=a[p];this.setAttribute(p,m.clone(t))}const l=e.morphAttributes;for(const p in l){const m=[],_=l[p];for(let x=0,S=_.length;x<S;x++)m.push(_[x].clone(t));this.morphAttributes[p]=m}this.morphTargetsRelative=e.morphTargetsRelative;const u=e.groups;for(let p=0,m=u.length;p<m;p++){const _=u[p];this.addGroup(_.start,_.count,_.materialIndex)}const f=e.boundingBox;f!==null&&(this.boundingBox=f.clone());const h=e.boundingSphere;return h!==null&&(this.boundingSphere=h.clone()),this.drawRange.start=e.drawRange.start,this.drawRange.count=e.drawRange.count,this.userData=e.userData,this}dispose(){this.dispatchEvent({type:"dispose"})}}const Um=new Gt,Xr=new Kg,wl=new Td,Fm=new Q,Al=new Q,Cl=new Q,Rl=new Q,uf=new Q,bl=new Q,Om=new Q,Pl=new Q;class xi extends vn{constructor(e=new ji,t=new Jg){super(),this.isMesh=!0,this.type="Mesh",this.geometry=e,this.material=t,this.updateMorphTargets()}copy(e,t){return super.copy(e,t),e.morphTargetInfluences!==void 0&&(this.morphTargetInfluences=e.morphTargetInfluences.slice()),e.morphTargetDictionary!==void 0&&(this.morphTargetDictionary=Object.assign({},e.morphTargetDictionary)),this.material=Array.isArray(e.material)?e.material.slice():e.material,this.geometry=e.geometry,this}updateMorphTargets(){const t=this.geometry.morphAttributes,s=Object.keys(t);if(s.length>0){const a=t[s[0]];if(a!==void 0){this.morphTargetInfluences=[],this.morphTargetDictionary={};for(let l=0,u=a.length;l<u;l++){const f=a[l].name||String(l);this.morphTargetInfluences.push(0),this.morphTargetDictionary[f]=l}}}}getVertexPosition(e,t){const s=this.geometry,a=s.attributes.position,l=s.morphAttributes.position,u=s.morphTargetsRelative;t.fromBufferAttribute(a,e);const f=this.morphTargetInfluences;if(l&&f){bl.set(0,0,0);for(let h=0,p=l.length;h<p;h++){const m=f[h],_=l[h];m!==0&&(uf.fromBufferAttribute(_,e),u?bl.addScaledVector(uf,m):bl.addScaledVector(uf.sub(t),m))}t.add(bl)}return t}raycast(e,t){const s=this.geometry,a=this.material,l=this.matrixWorld;a!==void 0&&(s.boundingSphere===null&&s.computeBoundingSphere(),wl.copy(s.boundingSphere),wl.applyMatrix4(l),Xr.copy(e.ray).recast(e.near),!(wl.containsPoint(Xr.origin)===!1&&(Xr.intersectSphere(wl,Fm)===null||Xr.origin.distanceToSquared(Fm)>(e.far-e.near)**2))&&(Um.copy(l).invert(),Xr.copy(e.ray).applyMatrix4(Um),!(s.boundingBox!==null&&Xr.intersectsBox(s.boundingBox)===!1)&&this._computeIntersections(e,t,Xr)))}_computeIntersections(e,t,s){let a;const l=this.geometry,u=this.material,f=l.index,h=l.attributes.position,p=l.attributes.uv,m=l.attributes.uv1,_=l.attributes.normal,x=l.groups,S=l.drawRange;if(f!==null)if(Array.isArray(u))for(let T=0,w=x.length;T<w;T++){const v=x[T],y=u[v.materialIndex],P=Math.max(v.start,S.start),b=Math.min(f.count,Math.min(v.start+v.count,S.start+S.count));for(let D=P,V=b;D<V;D+=3){const O=f.getX(D),U=f.getX(D+1),Y=f.getX(D+2);a=Ll(this,y,e,s,p,m,_,O,U,Y),a&&(a.faceIndex=Math.floor(D/3),a.face.materialIndex=v.materialIndex,t.push(a))}}else{const T=Math.max(0,S.start),w=Math.min(f.count,S.start+S.count);for(let v=T,y=w;v<y;v+=3){const P=f.getX(v),b=f.getX(v+1),D=f.getX(v+2);a=Ll(this,u,e,s,p,m,_,P,b,D),a&&(a.faceIndex=Math.floor(v/3),t.push(a))}}else if(h!==void 0)if(Array.isArray(u))for(let T=0,w=x.length;T<w;T++){const v=x[T],y=u[v.materialIndex],P=Math.max(v.start,S.start),b=Math.min(h.count,Math.min(v.start+v.count,S.start+S.count));for(let D=P,V=b;D<V;D+=3){const O=D,U=D+1,Y=D+2;a=Ll(this,y,e,s,p,m,_,O,U,Y),a&&(a.faceIndex=Math.floor(D/3),a.face.materialIndex=v.materialIndex,t.push(a))}}else{const T=Math.max(0,S.start),w=Math.min(h.count,S.start+S.count);for(let v=T,y=w;v<y;v+=3){const P=v,b=v+1,D=v+2;a=Ll(this,u,e,s,p,m,_,P,b,D),a&&(a.faceIndex=Math.floor(v/3),t.push(a))}}}}function Ey(r,e,t,s,a,l,u,f){let h;if(e.side===Ln?h=s.intersectTriangle(u,l,a,!0,f):h=s.intersectTriangle(a,l,u,e.side===Ar,f),h===null)return null;Pl.copy(f),Pl.applyMatrix4(r.matrixWorld);const p=t.ray.origin.distanceTo(Pl);return p<t.near||p>t.far?null:{distance:p,point:Pl.clone(),object:r}}function Ll(r,e,t,s,a,l,u,f,h,p){r.getVertexPosition(f,Al),r.getVertexPosition(h,Cl),r.getVertexPosition(p,Rl);const m=Ey(r,e,t,s,Al,Cl,Rl,Om);if(m){const _=new Q;fi.getBarycoord(Om,Al,Cl,Rl,_),a&&(m.uv=fi.getInterpolatedAttribute(a,f,h,p,_,new rt)),l&&(m.uv1=fi.getInterpolatedAttribute(l,f,h,p,_,new rt)),u&&(m.normal=fi.getInterpolatedAttribute(u,f,h,p,_,new Q),m.normal.dot(s.direction)>0&&m.normal.multiplyScalar(-1));const x={a:f,b:h,c:p,normal:new Q,materialIndex:0};fi.getNormal(Al,Cl,Rl,x.normal),m.face=x,m.barycoord=_}return m}class la extends ji{constructor(e=1,t=1,s=1,a=1,l=1,u=1){super(),this.type="BoxGeometry",this.parameters={width:e,height:t,depth:s,widthSegments:a,heightSegments:l,depthSegments:u};const f=this;a=Math.floor(a),l=Math.floor(l),u=Math.floor(u);const h=[],p=[],m=[],_=[];let x=0,S=0;T("z","y","x",-1,-1,s,t,e,u,l,0),T("z","y","x",1,-1,s,t,-e,u,l,1),T("x","z","y",1,1,e,s,t,a,u,2),T("x","z","y",1,-1,e,s,-t,a,u,3),T("x","y","z",1,-1,e,t,s,a,l,4),T("x","y","z",-1,-1,e,t,-s,a,l,5),this.setIndex(h),this.setAttribute("position",new Gi(p,3)),this.setAttribute("normal",new Gi(m,3)),this.setAttribute("uv",new Gi(_,2));function T(w,v,y,P,b,D,V,O,U,Y,ce){const E=D/U,C=V/Y,re=D/2,ee=V/2,ae=O/2,ue=U+1,Z=Y+1;let le=0,F=0;const se=new Q;for(let L=0;L<Z;L++){const X=L*C-ee;for(let ve=0;ve<ue;ve++){const Ne=ve*E-re;se[w]=Ne*P,se[v]=X*b,se[y]=ae,p.push(se.x,se.y,se.z),se[w]=0,se[v]=0,se[y]=O>0?1:-1,m.push(se.x,se.y,se.z),_.push(ve/U),_.push(1-L/Y),le+=1}}for(let L=0;L<Y;L++)for(let X=0;X<U;X++){const ve=x+X+ue*L,Ne=x+X+ue*(L+1),J=x+(X+1)+ue*(L+1),fe=x+(X+1)+ue*L;h.push(ve,Ne,fe),h.push(Ne,J,fe),F+=6}f.addGroup(S,F,ce),S+=F,x+=le}}copy(e){return super.copy(e),this.parameters=Object.assign({},e.parameters),this}static fromJSON(e){return new la(e.width,e.height,e.depth,e.widthSegments,e.heightSegments,e.depthSegments)}}function so(r){const e={};for(const t in r){e[t]={};for(const s in r[t]){const a=r[t][s];a&&(a.isColor||a.isMatrix3||a.isMatrix4||a.isVector2||a.isVector3||a.isVector4||a.isTexture||a.isQuaternion)?a.isRenderTargetTexture?(console.warn("UniformsUtils: Textures of render targets cannot be cloned via cloneUniforms() or mergeUniforms()."),e[t][s]=null):e[t][s]=a.clone():Array.isArray(a)?e[t][s]=a.slice():e[t][s]=a}}return e}function Sn(r){const e={};for(let t=0;t<r.length;t++){const s=so(r[t]);for(const a in s)e[a]=s[a]}return e}function Ty(r){const e=[];for(let t=0;t<r.length;t++)e.push(r[t].clone());return e}function n_(r){const e=r.getRenderTarget();return e===null?r.outputColorSpace:e.isXRRenderTarget===!0?e.texture.colorSpace:Tt.workingColorSpace}const wy={clone:so,merge:Sn};var Ay=`void main() {
	gl_Position = projectionMatrix * modelViewMatrix * vec4( position, 1.0 );
}`,Cy=`void main() {
	gl_FragColor = vec4( 1.0, 0.0, 0.0, 1.0 );
}`;class Cr extends aa{constructor(e){super(),this.isShaderMaterial=!0,this.type="ShaderMaterial",this.defines={},this.uniforms={},this.uniformsGroups=[],this.vertexShader=Ay,this.fragmentShader=Cy,this.linewidth=1,this.wireframe=!1,this.wireframeLinewidth=1,this.fog=!1,this.lights=!1,this.clipping=!1,this.forceSinglePass=!0,this.extensions={clipCullDistance:!1,multiDraw:!1},this.defaultAttributeValues={color:[1,1,1],uv:[0,0],uv1:[0,0]},this.index0AttributeName=void 0,this.uniformsNeedUpdate=!1,this.glslVersion=null,e!==void 0&&this.setValues(e)}copy(e){return super.copy(e),this.fragmentShader=e.fragmentShader,this.vertexShader=e.vertexShader,this.uniforms=so(e.uniforms),this.uniformsGroups=Ty(e.uniformsGroups),this.defines=Object.assign({},e.defines),this.wireframe=e.wireframe,this.wireframeLinewidth=e.wireframeLinewidth,this.fog=e.fog,this.lights=e.lights,this.clipping=e.clipping,this.extensions=Object.assign({},e.extensions),this.glslVersion=e.glslVersion,this}toJSON(e){const t=super.toJSON(e);t.glslVersion=this.glslVersion,t.uniforms={};for(const a in this.uniforms){const u=this.uniforms[a].value;u&&u.isTexture?t.uniforms[a]={type:"t",value:u.toJSON(e).uuid}:u&&u.isColor?t.uniforms[a]={type:"c",value:u.getHex()}:u&&u.isVector2?t.uniforms[a]={type:"v2",value:u.toArray()}:u&&u.isVector3?t.uniforms[a]={type:"v3",value:u.toArray()}:u&&u.isVector4?t.uniforms[a]={type:"v4",value:u.toArray()}:u&&u.isMatrix3?t.uniforms[a]={type:"m3",value:u.toArray()}:u&&u.isMatrix4?t.uniforms[a]={type:"m4",value:u.toArray()}:t.uniforms[a]={value:u}}Object.keys(this.defines).length>0&&(t.defines=this.defines),t.vertexShader=this.vertexShader,t.fragmentShader=this.fragmentShader,t.lights=this.lights,t.clipping=this.clipping;const s={};for(const a in this.extensions)this.extensions[a]===!0&&(s[a]=!0);return Object.keys(s).length>0&&(t.extensions=s),t}}class i_ extends vn{constructor(){super(),this.isCamera=!0,this.type="Camera",this.matrixWorldInverse=new Gt,this.projectionMatrix=new Gt,this.projectionMatrixInverse=new Gt,this.coordinateSystem=Vi}copy(e,t){return super.copy(e,t),this.matrixWorldInverse.copy(e.matrixWorldInverse),this.projectionMatrix.copy(e.projectionMatrix),this.projectionMatrixInverse.copy(e.projectionMatrixInverse),this.coordinateSystem=e.coordinateSystem,this}getWorldDirection(e){return super.getWorldDirection(e).negate()}updateMatrixWorld(e){super.updateMatrixWorld(e),this.matrixWorldInverse.copy(this.matrixWorld).invert()}updateWorldMatrix(e,t){super.updateWorldMatrix(e,t),this.matrixWorldInverse.copy(this.matrixWorld).invert()}clone(){return new this.constructor().copy(this)}}const xr=new Q,km=new rt,Bm=new rt;class Zn extends i_{constructor(e=50,t=1,s=.1,a=2e3){super(),this.isPerspectiveCamera=!0,this.type="PerspectiveCamera",this.fov=e,this.zoom=1,this.near=s,this.far=a,this.focus=10,this.aspect=t,this.view=null,this.filmGauge=35,this.filmOffset=0,this.updateProjectionMatrix()}copy(e,t){return super.copy(e,t),this.fov=e.fov,this.zoom=e.zoom,this.near=e.near,this.far=e.far,this.focus=e.focus,this.aspect=e.aspect,this.view=e.view===null?null:Object.assign({},e.view),this.filmGauge=e.filmGauge,this.filmOffset=e.filmOffset,this}setFocalLength(e){const t=.5*this.getFilmHeight()/e;this.fov=ud*2*Math.atan(t),this.updateProjectionMatrix()}getFocalLength(){const e=Math.tan(Gl*.5*this.fov);return .5*this.getFilmHeight()/e}getEffectiveFOV(){return ud*2*Math.atan(Math.tan(Gl*.5*this.fov)/this.zoom)}getFilmWidth(){return this.filmGauge*Math.min(this.aspect,1)}getFilmHeight(){return this.filmGauge/Math.max(this.aspect,1)}getViewBounds(e,t,s){xr.set(-1,-1,.5).applyMatrix4(this.projectionMatrixInverse),t.set(xr.x,xr.y).multiplyScalar(-e/xr.z),xr.set(1,1,.5).applyMatrix4(this.projectionMatrixInverse),s.set(xr.x,xr.y).multiplyScalar(-e/xr.z)}getViewSize(e,t){return this.getViewBounds(e,km,Bm),t.subVectors(Bm,km)}setViewOffset(e,t,s,a,l,u){this.aspect=e/t,this.view===null&&(this.view={enabled:!0,fullWidth:1,fullHeight:1,offsetX:0,offsetY:0,width:1,height:1}),this.view.enabled=!0,this.view.fullWidth=e,this.view.fullHeight=t,this.view.offsetX=s,this.view.offsetY=a,this.view.width=l,this.view.height=u,this.updateProjectionMatrix()}clearViewOffset(){this.view!==null&&(this.view.enabled=!1),this.updateProjectionMatrix()}updateProjectionMatrix(){const e=this.near;let t=e*Math.tan(Gl*.5*this.fov)/this.zoom,s=2*t,a=this.aspect*s,l=-.5*a;const u=this.view;if(this.view!==null&&this.view.enabled){const h=u.fullWidth,p=u.fullHeight;l+=u.offsetX*a/h,t-=u.offsetY*s/p,a*=u.width/h,s*=u.height/p}const f=this.filmOffset;f!==0&&(l+=e*f/this.getFilmWidth()),this.projectionMatrix.makePerspective(l,l+a,t,t-s,e,this.far,this.coordinateSystem),this.projectionMatrixInverse.copy(this.projectionMatrix).invert()}toJSON(e){const t=super.toJSON(e);return t.object.fov=this.fov,t.object.zoom=this.zoom,t.object.near=this.near,t.object.far=this.far,t.object.focus=this.focus,t.object.aspect=this.aspect,this.view!==null&&(t.object.view=Object.assign({},this.view)),t.object.filmGauge=this.filmGauge,t.object.filmOffset=this.filmOffset,t}}const js=-90,Xs=1;class Ry extends vn{constructor(e,t,s){super(),this.type="CubeCamera",this.renderTarget=s,this.coordinateSystem=null,this.activeMipmapLevel=0;const a=new Zn(js,Xs,e,t);a.layers=this.layers,this.add(a);const l=new Zn(js,Xs,e,t);l.layers=this.layers,this.add(l);const u=new Zn(js,Xs,e,t);u.layers=this.layers,this.add(u);const f=new Zn(js,Xs,e,t);f.layers=this.layers,this.add(f);const h=new Zn(js,Xs,e,t);h.layers=this.layers,this.add(h);const p=new Zn(js,Xs,e,t);p.layers=this.layers,this.add(p)}updateCoordinateSystem(){const e=this.coordinateSystem,t=this.children.concat(),[s,a,l,u,f,h]=t;for(const p of t)this.remove(p);if(e===Vi)s.up.set(0,1,0),s.lookAt(1,0,0),a.up.set(0,1,0),a.lookAt(-1,0,0),l.up.set(0,0,-1),l.lookAt(0,1,0),u.up.set(0,0,1),u.lookAt(0,-1,0),f.up.set(0,1,0),f.lookAt(0,0,1),h.up.set(0,1,0),h.lookAt(0,0,-1);else if(e===Kl)s.up.set(0,-1,0),s.lookAt(-1,0,0),a.up.set(0,-1,0),a.lookAt(1,0,0),l.up.set(0,0,1),l.lookAt(0,1,0),u.up.set(0,0,-1),u.lookAt(0,-1,0),f.up.set(0,-1,0),f.lookAt(0,0,1),h.up.set(0,-1,0),h.lookAt(0,0,-1);else throw new Error("THREE.CubeCamera.updateCoordinateSystem(): Invalid coordinate system: "+e);for(const p of t)this.add(p),p.updateMatrixWorld()}update(e,t){this.parent===null&&this.updateMatrixWorld();const{renderTarget:s,activeMipmapLevel:a}=this;this.coordinateSystem!==e.coordinateSystem&&(this.coordinateSystem=e.coordinateSystem,this.updateCoordinateSystem());const[l,u,f,h,p,m]=this.children,_=e.getRenderTarget(),x=e.getActiveCubeFace(),S=e.getActiveMipmapLevel(),T=e.xr.enabled;e.xr.enabled=!1;const w=s.texture.generateMipmaps;s.texture.generateMipmaps=!1,e.setRenderTarget(s,0,a),e.render(t,l),e.setRenderTarget(s,1,a),e.render(t,u),e.setRenderTarget(s,2,a),e.render(t,f),e.setRenderTarget(s,3,a),e.render(t,h),e.setRenderTarget(s,4,a),e.render(t,p),s.texture.generateMipmaps=w,e.setRenderTarget(s,5,a),e.render(t,m),e.setRenderTarget(_,x,S),e.xr.enabled=T,s.texture.needsPMREMUpdate=!0}}class r_ extends Dn{constructor(e,t,s,a,l,u,f,h,p,m){e=e!==void 0?e:[],t=t!==void 0?t:to,super(e,t,s,a,l,u,f,h,p,m),this.isCubeTexture=!0,this.flipY=!1}get images(){return this.image}set images(e){this.image=e}}class by extends is{constructor(e=1,t={}){super(e,e,t),this.isWebGLCubeRenderTarget=!0;const s={width:e,height:e,depth:1},a=[s,s,s,s,s,s];this.texture=new r_(a,t.mapping,t.wrapS,t.wrapT,t.magFilter,t.minFilter,t.format,t.type,t.anisotropy,t.colorSpace),this.texture.isRenderTargetTexture=!0,this.texture.generateMipmaps=t.generateMipmaps!==void 0?t.generateMipmaps:!1,this.texture.minFilter=t.minFilter!==void 0?t.minFilter:ui}fromEquirectangularTexture(e,t){this.texture.type=t.type,this.texture.colorSpace=t.colorSpace,this.texture.generateMipmaps=t.generateMipmaps,this.texture.minFilter=t.minFilter,this.texture.magFilter=t.magFilter;const s={uniforms:{tEquirect:{value:null}},vertexShader:`

				varying vec3 vWorldDirection;

				vec3 transformDirection( in vec3 dir, in mat4 matrix ) {

					return normalize( ( matrix * vec4( dir, 0.0 ) ).xyz );

				}

				void main() {

					vWorldDirection = transformDirection( position, modelMatrix );

					#include <begin_vertex>
					#include <project_vertex>

				}
			`,fragmentShader:`

				uniform sampler2D tEquirect;

				varying vec3 vWorldDirection;

				#include <common>

				void main() {

					vec3 direction = normalize( vWorldDirection );

					vec2 sampleUV = equirectUv( direction );

					gl_FragColor = texture2D( tEquirect, sampleUV );

				}
			`},a=new la(5,5,5),l=new Cr({name:"CubemapFromEquirect",uniforms:so(s.uniforms),vertexShader:s.vertexShader,fragmentShader:s.fragmentShader,side:Ln,blending:Tr});l.uniforms.tEquirect.value=t;const u=new xi(a,l),f=t.minFilter;return t.minFilter===ts&&(t.minFilter=ui),new Ry(1,10,this).update(e,u),t.minFilter=f,u.geometry.dispose(),u.material.dispose(),this}clear(e,t,s,a){const l=e.getRenderTarget();for(let u=0;u<6;u++)e.setRenderTarget(this,u),e.clear(t,s,a);e.setRenderTarget(l)}}const ff=new Q,Py=new Q,Ly=new ot;class yr{constructor(e=new Q(1,0,0),t=0){this.isPlane=!0,this.normal=e,this.constant=t}set(e,t){return this.normal.copy(e),this.constant=t,this}setComponents(e,t,s,a){return this.normal.set(e,t,s),this.constant=a,this}setFromNormalAndCoplanarPoint(e,t){return this.normal.copy(e),this.constant=-t.dot(this.normal),this}setFromCoplanarPoints(e,t,s){const a=ff.subVectors(s,t).cross(Py.subVectors(e,t)).normalize();return this.setFromNormalAndCoplanarPoint(a,e),this}copy(e){return this.normal.copy(e.normal),this.constant=e.constant,this}normalize(){const e=1/this.normal.length();return this.normal.multiplyScalar(e),this.constant*=e,this}negate(){return this.constant*=-1,this.normal.negate(),this}distanceToPoint(e){return this.normal.dot(e)+this.constant}distanceToSphere(e){return this.distanceToPoint(e.center)-e.radius}projectPoint(e,t){return t.copy(e).addScaledVector(this.normal,-this.distanceToPoint(e))}intersectLine(e,t){const s=e.delta(ff),a=this.normal.dot(s);if(a===0)return this.distanceToPoint(e.start)===0?t.copy(e.start):null;const l=-(e.start.dot(this.normal)+this.constant)/a;return l<0||l>1?null:t.copy(e.start).addScaledVector(s,l)}intersectsLine(e){const t=this.distanceToPoint(e.start),s=this.distanceToPoint(e.end);return t<0&&s>0||s<0&&t>0}intersectsBox(e){return e.intersectsPlane(this)}intersectsSphere(e){return e.intersectsPlane(this)}coplanarPoint(e){return e.copy(this.normal).multiplyScalar(-this.constant)}applyMatrix4(e,t){const s=t||Ly.getNormalMatrix(e),a=this.coplanarPoint(ff).applyMatrix4(e),l=this.normal.applyMatrix3(s).normalize();return this.constant=-a.dot(l),this}translate(e){return this.constant-=e.dot(this.normal),this}equals(e){return e.normal.equals(this.normal)&&e.constant===this.constant}clone(){return new this.constructor().copy(this)}}const Yr=new Td,Dl=new Q;class wd{constructor(e=new yr,t=new yr,s=new yr,a=new yr,l=new yr,u=new yr){this.planes=[e,t,s,a,l,u]}set(e,t,s,a,l,u){const f=this.planes;return f[0].copy(e),f[1].copy(t),f[2].copy(s),f[3].copy(a),f[4].copy(l),f[5].copy(u),this}copy(e){const t=this.planes;for(let s=0;s<6;s++)t[s].copy(e.planes[s]);return this}setFromProjectionMatrix(e,t=Vi){const s=this.planes,a=e.elements,l=a[0],u=a[1],f=a[2],h=a[3],p=a[4],m=a[5],_=a[6],x=a[7],S=a[8],T=a[9],w=a[10],v=a[11],y=a[12],P=a[13],b=a[14],D=a[15];if(s[0].setComponents(h-l,x-p,v-S,D-y).normalize(),s[1].setComponents(h+l,x+p,v+S,D+y).normalize(),s[2].setComponents(h+u,x+m,v+T,D+P).normalize(),s[3].setComponents(h-u,x-m,v-T,D-P).normalize(),s[4].setComponents(h-f,x-_,v-w,D-b).normalize(),t===Vi)s[5].setComponents(h+f,x+_,v+w,D+b).normalize();else if(t===Kl)s[5].setComponents(f,_,w,b).normalize();else throw new Error("THREE.Frustum.setFromProjectionMatrix(): Invalid coordinate system: "+t);return this}intersectsObject(e){if(e.boundingSphere!==void 0)e.boundingSphere===null&&e.computeBoundingSphere(),Yr.copy(e.boundingSphere).applyMatrix4(e.matrixWorld);else{const t=e.geometry;t.boundingSphere===null&&t.computeBoundingSphere(),Yr.copy(t.boundingSphere).applyMatrix4(e.matrixWorld)}return this.intersectsSphere(Yr)}intersectsSprite(e){return Yr.center.set(0,0,0),Yr.radius=.7071067811865476,Yr.applyMatrix4(e.matrixWorld),this.intersectsSphere(Yr)}intersectsSphere(e){const t=this.planes,s=e.center,a=-e.radius;for(let l=0;l<6;l++)if(t[l].distanceToPoint(s)<a)return!1;return!0}intersectsBox(e){const t=this.planes;for(let s=0;s<6;s++){const a=t[s];if(Dl.x=a.normal.x>0?e.max.x:e.min.x,Dl.y=a.normal.y>0?e.max.y:e.min.y,Dl.z=a.normal.z>0?e.max.z:e.min.z,a.distanceToPoint(Dl)<0)return!1}return!0}containsPoint(e){const t=this.planes;for(let s=0;s<6;s++)if(t[s].distanceToPoint(e)<0)return!1;return!0}clone(){return new this.constructor().copy(this)}}function s_(){let r=null,e=!1,t=null,s=null;function a(l,u){t(l,u),s=r.requestAnimationFrame(a)}return{start:function(){e!==!0&&t!==null&&(s=r.requestAnimationFrame(a),e=!0)},stop:function(){r.cancelAnimationFrame(s),e=!1},setAnimationLoop:function(l){t=l},setContext:function(l){r=l}}}function Dy(r){const e=new WeakMap;function t(f,h){const p=f.array,m=f.usage,_=p.byteLength,x=r.createBuffer();r.bindBuffer(h,x),r.bufferData(h,p,m),f.onUploadCallback();let S;if(p instanceof Float32Array)S=r.FLOAT;else if(p instanceof Uint16Array)f.isFloat16BufferAttribute?S=r.HALF_FLOAT:S=r.UNSIGNED_SHORT;else if(p instanceof Int16Array)S=r.SHORT;else if(p instanceof Uint32Array)S=r.UNSIGNED_INT;else if(p instanceof Int32Array)S=r.INT;else if(p instanceof Int8Array)S=r.BYTE;else if(p instanceof Uint8Array)S=r.UNSIGNED_BYTE;else if(p instanceof Uint8ClampedArray)S=r.UNSIGNED_BYTE;else throw new Error("THREE.WebGLAttributes: Unsupported buffer data format: "+p);return{buffer:x,type:S,bytesPerElement:p.BYTES_PER_ELEMENT,version:f.version,size:_}}function s(f,h,p){const m=h.array,_=h.updateRanges;if(r.bindBuffer(p,f),_.length===0)r.bufferSubData(p,0,m);else{_.sort((S,T)=>S.start-T.start);let x=0;for(let S=1;S<_.length;S++){const T=_[x],w=_[S];w.start<=T.start+T.count+1?T.count=Math.max(T.count,w.start+w.count-T.start):(++x,_[x]=w)}_.length=x+1;for(let S=0,T=_.length;S<T;S++){const w=_[S];r.bufferSubData(p,w.start*m.BYTES_PER_ELEMENT,m,w.start,w.count)}h.clearUpdateRanges()}h.onUploadCallback()}function a(f){return f.isInterleavedBufferAttribute&&(f=f.data),e.get(f)}function l(f){f.isInterleavedBufferAttribute&&(f=f.data);const h=e.get(f);h&&(r.deleteBuffer(h.buffer),e.delete(f))}function u(f,h){if(f.isInterleavedBufferAttribute&&(f=f.data),f.isGLBufferAttribute){const m=e.get(f);(!m||m.version<f.version)&&e.set(f,{buffer:f.buffer,type:f.type,bytesPerElement:f.elementSize,version:f.version});return}const p=e.get(f);if(p===void 0)e.set(f,t(f,h));else if(p.version<f.version){if(p.size!==f.array.byteLength)throw new Error("THREE.WebGLAttributes: The size of the buffer attribute's array buffer does not match the original size. Resizing buffer attributes is not supported.");s(p.buffer,f,h),p.version=f.version}}return{get:a,remove:l,update:u}}class ic extends ji{constructor(e=1,t=1,s=1,a=1){super(),this.type="PlaneGeometry",this.parameters={width:e,height:t,widthSegments:s,heightSegments:a};const l=e/2,u=t/2,f=Math.floor(s),h=Math.floor(a),p=f+1,m=h+1,_=e/f,x=t/h,S=[],T=[],w=[],v=[];for(let y=0;y<m;y++){const P=y*x-u;for(let b=0;b<p;b++){const D=b*_-l;T.push(D,-P,0),w.push(0,0,1),v.push(b/f),v.push(1-y/h)}}for(let y=0;y<h;y++)for(let P=0;P<f;P++){const b=P+p*y,D=P+p*(y+1),V=P+1+p*(y+1),O=P+1+p*y;S.push(b,D,O),S.push(D,V,O)}this.setIndex(S),this.setAttribute("position",new Gi(T,3)),this.setAttribute("normal",new Gi(w,3)),this.setAttribute("uv",new Gi(v,2))}copy(e){return super.copy(e),this.parameters=Object.assign({},e.parameters),this}static fromJSON(e){return new ic(e.width,e.height,e.widthSegments,e.heightSegments)}}var Ny=`#ifdef USE_ALPHAHASH
	if ( diffuseColor.a < getAlphaHashThreshold( vPosition ) ) discard;
#endif`,Iy=`#ifdef USE_ALPHAHASH
	const float ALPHA_HASH_SCALE = 0.05;
	float hash2D( vec2 value ) {
		return fract( 1.0e4 * sin( 17.0 * value.x + 0.1 * value.y ) * ( 0.1 + abs( sin( 13.0 * value.y + value.x ) ) ) );
	}
	float hash3D( vec3 value ) {
		return hash2D( vec2( hash2D( value.xy ), value.z ) );
	}
	float getAlphaHashThreshold( vec3 position ) {
		float maxDeriv = max(
			length( dFdx( position.xyz ) ),
			length( dFdy( position.xyz ) )
		);
		float pixScale = 1.0 / ( ALPHA_HASH_SCALE * maxDeriv );
		vec2 pixScales = vec2(
			exp2( floor( log2( pixScale ) ) ),
			exp2( ceil( log2( pixScale ) ) )
		);
		vec2 alpha = vec2(
			hash3D( floor( pixScales.x * position.xyz ) ),
			hash3D( floor( pixScales.y * position.xyz ) )
		);
		float lerpFactor = fract( log2( pixScale ) );
		float x = ( 1.0 - lerpFactor ) * alpha.x + lerpFactor * alpha.y;
		float a = min( lerpFactor, 1.0 - lerpFactor );
		vec3 cases = vec3(
			x * x / ( 2.0 * a * ( 1.0 - a ) ),
			( x - 0.5 * a ) / ( 1.0 - a ),
			1.0 - ( ( 1.0 - x ) * ( 1.0 - x ) / ( 2.0 * a * ( 1.0 - a ) ) )
		);
		float threshold = ( x < ( 1.0 - a ) )
			? ( ( x < a ) ? cases.x : cases.y )
			: cases.z;
		return clamp( threshold , 1.0e-6, 1.0 );
	}
#endif`,Uy=`#ifdef USE_ALPHAMAP
	diffuseColor.a *= texture2D( alphaMap, vAlphaMapUv ).g;
#endif`,Fy=`#ifdef USE_ALPHAMAP
	uniform sampler2D alphaMap;
#endif`,Oy=`#ifdef USE_ALPHATEST
	#ifdef ALPHA_TO_COVERAGE
	diffuseColor.a = smoothstep( alphaTest, alphaTest + fwidth( diffuseColor.a ), diffuseColor.a );
	if ( diffuseColor.a == 0.0 ) discard;
	#else
	if ( diffuseColor.a < alphaTest ) discard;
	#endif
#endif`,ky=`#ifdef USE_ALPHATEST
	uniform float alphaTest;
#endif`,By=`#ifdef USE_AOMAP
	float ambientOcclusion = ( texture2D( aoMap, vAoMapUv ).r - 1.0 ) * aoMapIntensity + 1.0;
	reflectedLight.indirectDiffuse *= ambientOcclusion;
	#if defined( USE_CLEARCOAT ) 
		clearcoatSpecularIndirect *= ambientOcclusion;
	#endif
	#if defined( USE_SHEEN ) 
		sheenSpecularIndirect *= ambientOcclusion;
	#endif
	#if defined( USE_ENVMAP ) && defined( STANDARD )
		float dotNV = saturate( dot( geometryNormal, geometryViewDir ) );
		reflectedLight.indirectSpecular *= computeSpecularOcclusion( dotNV, ambientOcclusion, material.roughness );
	#endif
#endif`,zy=`#ifdef USE_AOMAP
	uniform sampler2D aoMap;
	uniform float aoMapIntensity;
#endif`,Hy=`#ifdef USE_BATCHING
	#if ! defined( GL_ANGLE_multi_draw )
	#define gl_DrawID _gl_DrawID
	uniform int _gl_DrawID;
	#endif
	uniform highp sampler2D batchingTexture;
	uniform highp usampler2D batchingIdTexture;
	mat4 getBatchingMatrix( const in float i ) {
		int size = textureSize( batchingTexture, 0 ).x;
		int j = int( i ) * 4;
		int x = j % size;
		int y = j / size;
		vec4 v1 = texelFetch( batchingTexture, ivec2( x, y ), 0 );
		vec4 v2 = texelFetch( batchingTexture, ivec2( x + 1, y ), 0 );
		vec4 v3 = texelFetch( batchingTexture, ivec2( x + 2, y ), 0 );
		vec4 v4 = texelFetch( batchingTexture, ivec2( x + 3, y ), 0 );
		return mat4( v1, v2, v3, v4 );
	}
	float getIndirectIndex( const in int i ) {
		int size = textureSize( batchingIdTexture, 0 ).x;
		int x = i % size;
		int y = i / size;
		return float( texelFetch( batchingIdTexture, ivec2( x, y ), 0 ).r );
	}
#endif
#ifdef USE_BATCHING_COLOR
	uniform sampler2D batchingColorTexture;
	vec3 getBatchingColor( const in float i ) {
		int size = textureSize( batchingColorTexture, 0 ).x;
		int j = int( i );
		int x = j % size;
		int y = j / size;
		return texelFetch( batchingColorTexture, ivec2( x, y ), 0 ).rgb;
	}
#endif`,Vy=`#ifdef USE_BATCHING
	mat4 batchingMatrix = getBatchingMatrix( getIndirectIndex( gl_DrawID ) );
#endif`,Gy=`vec3 transformed = vec3( position );
#ifdef USE_ALPHAHASH
	vPosition = vec3( position );
#endif`,Wy=`vec3 objectNormal = vec3( normal );
#ifdef USE_TANGENT
	vec3 objectTangent = vec3( tangent.xyz );
#endif`,jy=`float G_BlinnPhong_Implicit( ) {
	return 0.25;
}
float D_BlinnPhong( const in float shininess, const in float dotNH ) {
	return RECIPROCAL_PI * ( shininess * 0.5 + 1.0 ) * pow( dotNH, shininess );
}
vec3 BRDF_BlinnPhong( const in vec3 lightDir, const in vec3 viewDir, const in vec3 normal, const in vec3 specularColor, const in float shininess ) {
	vec3 halfDir = normalize( lightDir + viewDir );
	float dotNH = saturate( dot( normal, halfDir ) );
	float dotVH = saturate( dot( viewDir, halfDir ) );
	vec3 F = F_Schlick( specularColor, 1.0, dotVH );
	float G = G_BlinnPhong_Implicit( );
	float D = D_BlinnPhong( shininess, dotNH );
	return F * ( G * D );
} // validated`,Xy=`#ifdef USE_IRIDESCENCE
	const mat3 XYZ_TO_REC709 = mat3(
		 3.2404542, -0.9692660,  0.0556434,
		-1.5371385,  1.8760108, -0.2040259,
		-0.4985314,  0.0415560,  1.0572252
	);
	vec3 Fresnel0ToIor( vec3 fresnel0 ) {
		vec3 sqrtF0 = sqrt( fresnel0 );
		return ( vec3( 1.0 ) + sqrtF0 ) / ( vec3( 1.0 ) - sqrtF0 );
	}
	vec3 IorToFresnel0( vec3 transmittedIor, float incidentIor ) {
		return pow2( ( transmittedIor - vec3( incidentIor ) ) / ( transmittedIor + vec3( incidentIor ) ) );
	}
	float IorToFresnel0( float transmittedIor, float incidentIor ) {
		return pow2( ( transmittedIor - incidentIor ) / ( transmittedIor + incidentIor ));
	}
	vec3 evalSensitivity( float OPD, vec3 shift ) {
		float phase = 2.0 * PI * OPD * 1.0e-9;
		vec3 val = vec3( 5.4856e-13, 4.4201e-13, 5.2481e-13 );
		vec3 pos = vec3( 1.6810e+06, 1.7953e+06, 2.2084e+06 );
		vec3 var = vec3( 4.3278e+09, 9.3046e+09, 6.6121e+09 );
		vec3 xyz = val * sqrt( 2.0 * PI * var ) * cos( pos * phase + shift ) * exp( - pow2( phase ) * var );
		xyz.x += 9.7470e-14 * sqrt( 2.0 * PI * 4.5282e+09 ) * cos( 2.2399e+06 * phase + shift[ 0 ] ) * exp( - 4.5282e+09 * pow2( phase ) );
		xyz /= 1.0685e-7;
		vec3 rgb = XYZ_TO_REC709 * xyz;
		return rgb;
	}
	vec3 evalIridescence( float outsideIOR, float eta2, float cosTheta1, float thinFilmThickness, vec3 baseF0 ) {
		vec3 I;
		float iridescenceIOR = mix( outsideIOR, eta2, smoothstep( 0.0, 0.03, thinFilmThickness ) );
		float sinTheta2Sq = pow2( outsideIOR / iridescenceIOR ) * ( 1.0 - pow2( cosTheta1 ) );
		float cosTheta2Sq = 1.0 - sinTheta2Sq;
		if ( cosTheta2Sq < 0.0 ) {
			return vec3( 1.0 );
		}
		float cosTheta2 = sqrt( cosTheta2Sq );
		float R0 = IorToFresnel0( iridescenceIOR, outsideIOR );
		float R12 = F_Schlick( R0, 1.0, cosTheta1 );
		float T121 = 1.0 - R12;
		float phi12 = 0.0;
		if ( iridescenceIOR < outsideIOR ) phi12 = PI;
		float phi21 = PI - phi12;
		vec3 baseIOR = Fresnel0ToIor( clamp( baseF0, 0.0, 0.9999 ) );		vec3 R1 = IorToFresnel0( baseIOR, iridescenceIOR );
		vec3 R23 = F_Schlick( R1, 1.0, cosTheta2 );
		vec3 phi23 = vec3( 0.0 );
		if ( baseIOR[ 0 ] < iridescenceIOR ) phi23[ 0 ] = PI;
		if ( baseIOR[ 1 ] < iridescenceIOR ) phi23[ 1 ] = PI;
		if ( baseIOR[ 2 ] < iridescenceIOR ) phi23[ 2 ] = PI;
		float OPD = 2.0 * iridescenceIOR * thinFilmThickness * cosTheta2;
		vec3 phi = vec3( phi21 ) + phi23;
		vec3 R123 = clamp( R12 * R23, 1e-5, 0.9999 );
		vec3 r123 = sqrt( R123 );
		vec3 Rs = pow2( T121 ) * R23 / ( vec3( 1.0 ) - R123 );
		vec3 C0 = R12 + Rs;
		I = C0;
		vec3 Cm = Rs - T121;
		for ( int m = 1; m <= 2; ++ m ) {
			Cm *= r123;
			vec3 Sm = 2.0 * evalSensitivity( float( m ) * OPD, float( m ) * phi );
			I += Cm * Sm;
		}
		return max( I, vec3( 0.0 ) );
	}
#endif`,Yy=`#ifdef USE_BUMPMAP
	uniform sampler2D bumpMap;
	uniform float bumpScale;
	vec2 dHdxy_fwd() {
		vec2 dSTdx = dFdx( vBumpMapUv );
		vec2 dSTdy = dFdy( vBumpMapUv );
		float Hll = bumpScale * texture2D( bumpMap, vBumpMapUv ).x;
		float dBx = bumpScale * texture2D( bumpMap, vBumpMapUv + dSTdx ).x - Hll;
		float dBy = bumpScale * texture2D( bumpMap, vBumpMapUv + dSTdy ).x - Hll;
		return vec2( dBx, dBy );
	}
	vec3 perturbNormalArb( vec3 surf_pos, vec3 surf_norm, vec2 dHdxy, float faceDirection ) {
		vec3 vSigmaX = normalize( dFdx( surf_pos.xyz ) );
		vec3 vSigmaY = normalize( dFdy( surf_pos.xyz ) );
		vec3 vN = surf_norm;
		vec3 R1 = cross( vSigmaY, vN );
		vec3 R2 = cross( vN, vSigmaX );
		float fDet = dot( vSigmaX, R1 ) * faceDirection;
		vec3 vGrad = sign( fDet ) * ( dHdxy.x * R1 + dHdxy.y * R2 );
		return normalize( abs( fDet ) * surf_norm - vGrad );
	}
#endif`,qy=`#if NUM_CLIPPING_PLANES > 0
	vec4 plane;
	#ifdef ALPHA_TO_COVERAGE
		float distanceToPlane, distanceGradient;
		float clipOpacity = 1.0;
		#pragma unroll_loop_start
		for ( int i = 0; i < UNION_CLIPPING_PLANES; i ++ ) {
			plane = clippingPlanes[ i ];
			distanceToPlane = - dot( vClipPosition, plane.xyz ) + plane.w;
			distanceGradient = fwidth( distanceToPlane ) / 2.0;
			clipOpacity *= smoothstep( - distanceGradient, distanceGradient, distanceToPlane );
			if ( clipOpacity == 0.0 ) discard;
		}
		#pragma unroll_loop_end
		#if UNION_CLIPPING_PLANES < NUM_CLIPPING_PLANES
			float unionClipOpacity = 1.0;
			#pragma unroll_loop_start
			for ( int i = UNION_CLIPPING_PLANES; i < NUM_CLIPPING_PLANES; i ++ ) {
				plane = clippingPlanes[ i ];
				distanceToPlane = - dot( vClipPosition, plane.xyz ) + plane.w;
				distanceGradient = fwidth( distanceToPlane ) / 2.0;
				unionClipOpacity *= 1.0 - smoothstep( - distanceGradient, distanceGradient, distanceToPlane );
			}
			#pragma unroll_loop_end
			clipOpacity *= 1.0 - unionClipOpacity;
		#endif
		diffuseColor.a *= clipOpacity;
		if ( diffuseColor.a == 0.0 ) discard;
	#else
		#pragma unroll_loop_start
		for ( int i = 0; i < UNION_CLIPPING_PLANES; i ++ ) {
			plane = clippingPlanes[ i ];
			if ( dot( vClipPosition, plane.xyz ) > plane.w ) discard;
		}
		#pragma unroll_loop_end
		#if UNION_CLIPPING_PLANES < NUM_CLIPPING_PLANES
			bool clipped = true;
			#pragma unroll_loop_start
			for ( int i = UNION_CLIPPING_PLANES; i < NUM_CLIPPING_PLANES; i ++ ) {
				plane = clippingPlanes[ i ];
				clipped = ( dot( vClipPosition, plane.xyz ) > plane.w ) && clipped;
			}
			#pragma unroll_loop_end
			if ( clipped ) discard;
		#endif
	#endif
#endif`,$y=`#if NUM_CLIPPING_PLANES > 0
	varying vec3 vClipPosition;
	uniform vec4 clippingPlanes[ NUM_CLIPPING_PLANES ];
#endif`,Ky=`#if NUM_CLIPPING_PLANES > 0
	varying vec3 vClipPosition;
#endif`,Zy=`#if NUM_CLIPPING_PLANES > 0
	vClipPosition = - mvPosition.xyz;
#endif`,Qy=`#if defined( USE_COLOR_ALPHA )
	diffuseColor *= vColor;
#elif defined( USE_COLOR )
	diffuseColor.rgb *= vColor;
#endif`,Jy=`#if defined( USE_COLOR_ALPHA )
	varying vec4 vColor;
#elif defined( USE_COLOR )
	varying vec3 vColor;
#endif`,eS=`#if defined( USE_COLOR_ALPHA )
	varying vec4 vColor;
#elif defined( USE_COLOR ) || defined( USE_INSTANCING_COLOR ) || defined( USE_BATCHING_COLOR )
	varying vec3 vColor;
#endif`,tS=`#if defined( USE_COLOR_ALPHA )
	vColor = vec4( 1.0 );
#elif defined( USE_COLOR ) || defined( USE_INSTANCING_COLOR ) || defined( USE_BATCHING_COLOR )
	vColor = vec3( 1.0 );
#endif
#ifdef USE_COLOR
	vColor *= color;
#endif
#ifdef USE_INSTANCING_COLOR
	vColor.xyz *= instanceColor.xyz;
#endif
#ifdef USE_BATCHING_COLOR
	vec3 batchingColor = getBatchingColor( getIndirectIndex( gl_DrawID ) );
	vColor.xyz *= batchingColor.xyz;
#endif`,nS=`#define PI 3.141592653589793
#define PI2 6.283185307179586
#define PI_HALF 1.5707963267948966
#define RECIPROCAL_PI 0.3183098861837907
#define RECIPROCAL_PI2 0.15915494309189535
#define EPSILON 1e-6
#ifndef saturate
#define saturate( a ) clamp( a, 0.0, 1.0 )
#endif
#define whiteComplement( a ) ( 1.0 - saturate( a ) )
float pow2( const in float x ) { return x*x; }
vec3 pow2( const in vec3 x ) { return x*x; }
float pow3( const in float x ) { return x*x*x; }
float pow4( const in float x ) { float x2 = x*x; return x2*x2; }
float max3( const in vec3 v ) { return max( max( v.x, v.y ), v.z ); }
float average( const in vec3 v ) { return dot( v, vec3( 0.3333333 ) ); }
highp float rand( const in vec2 uv ) {
	const highp float a = 12.9898, b = 78.233, c = 43758.5453;
	highp float dt = dot( uv.xy, vec2( a,b ) ), sn = mod( dt, PI );
	return fract( sin( sn ) * c );
}
#ifdef HIGH_PRECISION
	float precisionSafeLength( vec3 v ) { return length( v ); }
#else
	float precisionSafeLength( vec3 v ) {
		float maxComponent = max3( abs( v ) );
		return length( v / maxComponent ) * maxComponent;
	}
#endif
struct IncidentLight {
	vec3 color;
	vec3 direction;
	bool visible;
};
struct ReflectedLight {
	vec3 directDiffuse;
	vec3 directSpecular;
	vec3 indirectDiffuse;
	vec3 indirectSpecular;
};
#ifdef USE_ALPHAHASH
	varying vec3 vPosition;
#endif
vec3 transformDirection( in vec3 dir, in mat4 matrix ) {
	return normalize( ( matrix * vec4( dir, 0.0 ) ).xyz );
}
vec3 inverseTransformDirection( in vec3 dir, in mat4 matrix ) {
	return normalize( ( vec4( dir, 0.0 ) * matrix ).xyz );
}
mat3 transposeMat3( const in mat3 m ) {
	mat3 tmp;
	tmp[ 0 ] = vec3( m[ 0 ].x, m[ 1 ].x, m[ 2 ].x );
	tmp[ 1 ] = vec3( m[ 0 ].y, m[ 1 ].y, m[ 2 ].y );
	tmp[ 2 ] = vec3( m[ 0 ].z, m[ 1 ].z, m[ 2 ].z );
	return tmp;
}
bool isPerspectiveMatrix( mat4 m ) {
	return m[ 2 ][ 3 ] == - 1.0;
}
vec2 equirectUv( in vec3 dir ) {
	float u = atan( dir.z, dir.x ) * RECIPROCAL_PI2 + 0.5;
	float v = asin( clamp( dir.y, - 1.0, 1.0 ) ) * RECIPROCAL_PI + 0.5;
	return vec2( u, v );
}
vec3 BRDF_Lambert( const in vec3 diffuseColor ) {
	return RECIPROCAL_PI * diffuseColor;
}
vec3 F_Schlick( const in vec3 f0, const in float f90, const in float dotVH ) {
	float fresnel = exp2( ( - 5.55473 * dotVH - 6.98316 ) * dotVH );
	return f0 * ( 1.0 - fresnel ) + ( f90 * fresnel );
}
float F_Schlick( const in float f0, const in float f90, const in float dotVH ) {
	float fresnel = exp2( ( - 5.55473 * dotVH - 6.98316 ) * dotVH );
	return f0 * ( 1.0 - fresnel ) + ( f90 * fresnel );
} // validated`,iS=`#ifdef ENVMAP_TYPE_CUBE_UV
	#define cubeUV_minMipLevel 4.0
	#define cubeUV_minTileSize 16.0
	float getFace( vec3 direction ) {
		vec3 absDirection = abs( direction );
		float face = - 1.0;
		if ( absDirection.x > absDirection.z ) {
			if ( absDirection.x > absDirection.y )
				face = direction.x > 0.0 ? 0.0 : 3.0;
			else
				face = direction.y > 0.0 ? 1.0 : 4.0;
		} else {
			if ( absDirection.z > absDirection.y )
				face = direction.z > 0.0 ? 2.0 : 5.0;
			else
				face = direction.y > 0.0 ? 1.0 : 4.0;
		}
		return face;
	}
	vec2 getUV( vec3 direction, float face ) {
		vec2 uv;
		if ( face == 0.0 ) {
			uv = vec2( direction.z, direction.y ) / abs( direction.x );
		} else if ( face == 1.0 ) {
			uv = vec2( - direction.x, - direction.z ) / abs( direction.y );
		} else if ( face == 2.0 ) {
			uv = vec2( - direction.x, direction.y ) / abs( direction.z );
		} else if ( face == 3.0 ) {
			uv = vec2( - direction.z, direction.y ) / abs( direction.x );
		} else if ( face == 4.0 ) {
			uv = vec2( - direction.x, direction.z ) / abs( direction.y );
		} else {
			uv = vec2( direction.x, direction.y ) / abs( direction.z );
		}
		return 0.5 * ( uv + 1.0 );
	}
	vec3 bilinearCubeUV( sampler2D envMap, vec3 direction, float mipInt ) {
		float face = getFace( direction );
		float filterInt = max( cubeUV_minMipLevel - mipInt, 0.0 );
		mipInt = max( mipInt, cubeUV_minMipLevel );
		float faceSize = exp2( mipInt );
		highp vec2 uv = getUV( direction, face ) * ( faceSize - 2.0 ) + 1.0;
		if ( face > 2.0 ) {
			uv.y += faceSize;
			face -= 3.0;
		}
		uv.x += face * faceSize;
		uv.x += filterInt * 3.0 * cubeUV_minTileSize;
		uv.y += 4.0 * ( exp2( CUBEUV_MAX_MIP ) - faceSize );
		uv.x *= CUBEUV_TEXEL_WIDTH;
		uv.y *= CUBEUV_TEXEL_HEIGHT;
		#ifdef texture2DGradEXT
			return texture2DGradEXT( envMap, uv, vec2( 0.0 ), vec2( 0.0 ) ).rgb;
		#else
			return texture2D( envMap, uv ).rgb;
		#endif
	}
	#define cubeUV_r0 1.0
	#define cubeUV_m0 - 2.0
	#define cubeUV_r1 0.8
	#define cubeUV_m1 - 1.0
	#define cubeUV_r4 0.4
	#define cubeUV_m4 2.0
	#define cubeUV_r5 0.305
	#define cubeUV_m5 3.0
	#define cubeUV_r6 0.21
	#define cubeUV_m6 4.0
	float roughnessToMip( float roughness ) {
		float mip = 0.0;
		if ( roughness >= cubeUV_r1 ) {
			mip = ( cubeUV_r0 - roughness ) * ( cubeUV_m1 - cubeUV_m0 ) / ( cubeUV_r0 - cubeUV_r1 ) + cubeUV_m0;
		} else if ( roughness >= cubeUV_r4 ) {
			mip = ( cubeUV_r1 - roughness ) * ( cubeUV_m4 - cubeUV_m1 ) / ( cubeUV_r1 - cubeUV_r4 ) + cubeUV_m1;
		} else if ( roughness >= cubeUV_r5 ) {
			mip = ( cubeUV_r4 - roughness ) * ( cubeUV_m5 - cubeUV_m4 ) / ( cubeUV_r4 - cubeUV_r5 ) + cubeUV_m4;
		} else if ( roughness >= cubeUV_r6 ) {
			mip = ( cubeUV_r5 - roughness ) * ( cubeUV_m6 - cubeUV_m5 ) / ( cubeUV_r5 - cubeUV_r6 ) + cubeUV_m5;
		} else {
			mip = - 2.0 * log2( 1.16 * roughness );		}
		return mip;
	}
	vec4 textureCubeUV( sampler2D envMap, vec3 sampleDir, float roughness ) {
		float mip = clamp( roughnessToMip( roughness ), cubeUV_m0, CUBEUV_MAX_MIP );
		float mipF = fract( mip );
		float mipInt = floor( mip );
		vec3 color0 = bilinearCubeUV( envMap, sampleDir, mipInt );
		if ( mipF == 0.0 ) {
			return vec4( color0, 1.0 );
		} else {
			vec3 color1 = bilinearCubeUV( envMap, sampleDir, mipInt + 1.0 );
			return vec4( mix( color0, color1, mipF ), 1.0 );
		}
	}
#endif`,rS=`vec3 transformedNormal = objectNormal;
#ifdef USE_TANGENT
	vec3 transformedTangent = objectTangent;
#endif
#ifdef USE_BATCHING
	mat3 bm = mat3( batchingMatrix );
	transformedNormal /= vec3( dot( bm[ 0 ], bm[ 0 ] ), dot( bm[ 1 ], bm[ 1 ] ), dot( bm[ 2 ], bm[ 2 ] ) );
	transformedNormal = bm * transformedNormal;
	#ifdef USE_TANGENT
		transformedTangent = bm * transformedTangent;
	#endif
#endif
#ifdef USE_INSTANCING
	mat3 im = mat3( instanceMatrix );
	transformedNormal /= vec3( dot( im[ 0 ], im[ 0 ] ), dot( im[ 1 ], im[ 1 ] ), dot( im[ 2 ], im[ 2 ] ) );
	transformedNormal = im * transformedNormal;
	#ifdef USE_TANGENT
		transformedTangent = im * transformedTangent;
	#endif
#endif
transformedNormal = normalMatrix * transformedNormal;
#ifdef FLIP_SIDED
	transformedNormal = - transformedNormal;
#endif
#ifdef USE_TANGENT
	transformedTangent = ( modelViewMatrix * vec4( transformedTangent, 0.0 ) ).xyz;
	#ifdef FLIP_SIDED
		transformedTangent = - transformedTangent;
	#endif
#endif`,sS=`#ifdef USE_DISPLACEMENTMAP
	uniform sampler2D displacementMap;
	uniform float displacementScale;
	uniform float displacementBias;
#endif`,oS=`#ifdef USE_DISPLACEMENTMAP
	transformed += normalize( objectNormal ) * ( texture2D( displacementMap, vDisplacementMapUv ).x * displacementScale + displacementBias );
#endif`,aS=`#ifdef USE_EMISSIVEMAP
	vec4 emissiveColor = texture2D( emissiveMap, vEmissiveMapUv );
	totalEmissiveRadiance *= emissiveColor.rgb;
#endif`,lS=`#ifdef USE_EMISSIVEMAP
	uniform sampler2D emissiveMap;
#endif`,cS="gl_FragColor = linearToOutputTexel( gl_FragColor );",uS=`
const mat3 LINEAR_SRGB_TO_LINEAR_DISPLAY_P3 = mat3(
	vec3( 0.8224621, 0.177538, 0.0 ),
	vec3( 0.0331941, 0.9668058, 0.0 ),
	vec3( 0.0170827, 0.0723974, 0.9105199 )
);
const mat3 LINEAR_DISPLAY_P3_TO_LINEAR_SRGB = mat3(
	vec3( 1.2249401, - 0.2249404, 0.0 ),
	vec3( - 0.0420569, 1.0420571, 0.0 ),
	vec3( - 0.0196376, - 0.0786361, 1.0982735 )
);
vec4 LinearSRGBToLinearDisplayP3( in vec4 value ) {
	return vec4( value.rgb * LINEAR_SRGB_TO_LINEAR_DISPLAY_P3, value.a );
}
vec4 LinearDisplayP3ToLinearSRGB( in vec4 value ) {
	return vec4( value.rgb * LINEAR_DISPLAY_P3_TO_LINEAR_SRGB, value.a );
}
vec4 LinearTransferOETF( in vec4 value ) {
	return value;
}
vec4 sRGBTransferOETF( in vec4 value ) {
	return vec4( mix( pow( value.rgb, vec3( 0.41666 ) ) * 1.055 - vec3( 0.055 ), value.rgb * 12.92, vec3( lessThanEqual( value.rgb, vec3( 0.0031308 ) ) ) ), value.a );
}`,fS=`#ifdef USE_ENVMAP
	#ifdef ENV_WORLDPOS
		vec3 cameraToFrag;
		if ( isOrthographic ) {
			cameraToFrag = normalize( vec3( - viewMatrix[ 0 ][ 2 ], - viewMatrix[ 1 ][ 2 ], - viewMatrix[ 2 ][ 2 ] ) );
		} else {
			cameraToFrag = normalize( vWorldPosition - cameraPosition );
		}
		vec3 worldNormal = inverseTransformDirection( normal, viewMatrix );
		#ifdef ENVMAP_MODE_REFLECTION
			vec3 reflectVec = reflect( cameraToFrag, worldNormal );
		#else
			vec3 reflectVec = refract( cameraToFrag, worldNormal, refractionRatio );
		#endif
	#else
		vec3 reflectVec = vReflect;
	#endif
	#ifdef ENVMAP_TYPE_CUBE
		vec4 envColor = textureCube( envMap, envMapRotation * vec3( flipEnvMap * reflectVec.x, reflectVec.yz ) );
	#else
		vec4 envColor = vec4( 0.0 );
	#endif
	#ifdef ENVMAP_BLENDING_MULTIPLY
		outgoingLight = mix( outgoingLight, outgoingLight * envColor.xyz, specularStrength * reflectivity );
	#elif defined( ENVMAP_BLENDING_MIX )
		outgoingLight = mix( outgoingLight, envColor.xyz, specularStrength * reflectivity );
	#elif defined( ENVMAP_BLENDING_ADD )
		outgoingLight += envColor.xyz * specularStrength * reflectivity;
	#endif
#endif`,dS=`#ifdef USE_ENVMAP
	uniform float envMapIntensity;
	uniform float flipEnvMap;
	uniform mat3 envMapRotation;
	#ifdef ENVMAP_TYPE_CUBE
		uniform samplerCube envMap;
	#else
		uniform sampler2D envMap;
	#endif
	
#endif`,hS=`#ifdef USE_ENVMAP
	uniform float reflectivity;
	#if defined( USE_BUMPMAP ) || defined( USE_NORMALMAP ) || defined( PHONG ) || defined( LAMBERT )
		#define ENV_WORLDPOS
	#endif
	#ifdef ENV_WORLDPOS
		varying vec3 vWorldPosition;
		uniform float refractionRatio;
	#else
		varying vec3 vReflect;
	#endif
#endif`,pS=`#ifdef USE_ENVMAP
	#if defined( USE_BUMPMAP ) || defined( USE_NORMALMAP ) || defined( PHONG ) || defined( LAMBERT )
		#define ENV_WORLDPOS
	#endif
	#ifdef ENV_WORLDPOS
		
		varying vec3 vWorldPosition;
	#else
		varying vec3 vReflect;
		uniform float refractionRatio;
	#endif
#endif`,mS=`#ifdef USE_ENVMAP
	#ifdef ENV_WORLDPOS
		vWorldPosition = worldPosition.xyz;
	#else
		vec3 cameraToVertex;
		if ( isOrthographic ) {
			cameraToVertex = normalize( vec3( - viewMatrix[ 0 ][ 2 ], - viewMatrix[ 1 ][ 2 ], - viewMatrix[ 2 ][ 2 ] ) );
		} else {
			cameraToVertex = normalize( worldPosition.xyz - cameraPosition );
		}
		vec3 worldNormal = inverseTransformDirection( transformedNormal, viewMatrix );
		#ifdef ENVMAP_MODE_REFLECTION
			vReflect = reflect( cameraToVertex, worldNormal );
		#else
			vReflect = refract( cameraToVertex, worldNormal, refractionRatio );
		#endif
	#endif
#endif`,gS=`#ifdef USE_FOG
	vFogDepth = - mvPosition.z;
#endif`,_S=`#ifdef USE_FOG
	varying float vFogDepth;
#endif`,vS=`#ifdef USE_FOG
	#ifdef FOG_EXP2
		float fogFactor = 1.0 - exp( - fogDensity * fogDensity * vFogDepth * vFogDepth );
	#else
		float fogFactor = smoothstep( fogNear, fogFar, vFogDepth );
	#endif
	gl_FragColor.rgb = mix( gl_FragColor.rgb, fogColor, fogFactor );
#endif`,xS=`#ifdef USE_FOG
	uniform vec3 fogColor;
	varying float vFogDepth;
	#ifdef FOG_EXP2
		uniform float fogDensity;
	#else
		uniform float fogNear;
		uniform float fogFar;
	#endif
#endif`,yS=`#ifdef USE_GRADIENTMAP
	uniform sampler2D gradientMap;
#endif
vec3 getGradientIrradiance( vec3 normal, vec3 lightDirection ) {
	float dotNL = dot( normal, lightDirection );
	vec2 coord = vec2( dotNL * 0.5 + 0.5, 0.0 );
	#ifdef USE_GRADIENTMAP
		return vec3( texture2D( gradientMap, coord ).r );
	#else
		vec2 fw = fwidth( coord ) * 0.5;
		return mix( vec3( 0.7 ), vec3( 1.0 ), smoothstep( 0.7 - fw.x, 0.7 + fw.x, coord.x ) );
	#endif
}`,SS=`#ifdef USE_LIGHTMAP
	uniform sampler2D lightMap;
	uniform float lightMapIntensity;
#endif`,MS=`LambertMaterial material;
material.diffuseColor = diffuseColor.rgb;
material.specularStrength = specularStrength;`,ES=`varying vec3 vViewPosition;
struct LambertMaterial {
	vec3 diffuseColor;
	float specularStrength;
};
void RE_Direct_Lambert( const in IncidentLight directLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in LambertMaterial material, inout ReflectedLight reflectedLight ) {
	float dotNL = saturate( dot( geometryNormal, directLight.direction ) );
	vec3 irradiance = dotNL * directLight.color;
	reflectedLight.directDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
void RE_IndirectDiffuse_Lambert( const in vec3 irradiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in LambertMaterial material, inout ReflectedLight reflectedLight ) {
	reflectedLight.indirectDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
#define RE_Direct				RE_Direct_Lambert
#define RE_IndirectDiffuse		RE_IndirectDiffuse_Lambert`,TS=`uniform bool receiveShadow;
uniform vec3 ambientLightColor;
#if defined( USE_LIGHT_PROBES )
	uniform vec3 lightProbe[ 9 ];
#endif
vec3 shGetIrradianceAt( in vec3 normal, in vec3 shCoefficients[ 9 ] ) {
	float x = normal.x, y = normal.y, z = normal.z;
	vec3 result = shCoefficients[ 0 ] * 0.886227;
	result += shCoefficients[ 1 ] * 2.0 * 0.511664 * y;
	result += shCoefficients[ 2 ] * 2.0 * 0.511664 * z;
	result += shCoefficients[ 3 ] * 2.0 * 0.511664 * x;
	result += shCoefficients[ 4 ] * 2.0 * 0.429043 * x * y;
	result += shCoefficients[ 5 ] * 2.0 * 0.429043 * y * z;
	result += shCoefficients[ 6 ] * ( 0.743125 * z * z - 0.247708 );
	result += shCoefficients[ 7 ] * 2.0 * 0.429043 * x * z;
	result += shCoefficients[ 8 ] * 0.429043 * ( x * x - y * y );
	return result;
}
vec3 getLightProbeIrradiance( const in vec3 lightProbe[ 9 ], const in vec3 normal ) {
	vec3 worldNormal = inverseTransformDirection( normal, viewMatrix );
	vec3 irradiance = shGetIrradianceAt( worldNormal, lightProbe );
	return irradiance;
}
vec3 getAmbientLightIrradiance( const in vec3 ambientLightColor ) {
	vec3 irradiance = ambientLightColor;
	return irradiance;
}
float getDistanceAttenuation( const in float lightDistance, const in float cutoffDistance, const in float decayExponent ) {
	float distanceFalloff = 1.0 / max( pow( lightDistance, decayExponent ), 0.01 );
	if ( cutoffDistance > 0.0 ) {
		distanceFalloff *= pow2( saturate( 1.0 - pow4( lightDistance / cutoffDistance ) ) );
	}
	return distanceFalloff;
}
float getSpotAttenuation( const in float coneCosine, const in float penumbraCosine, const in float angleCosine ) {
	return smoothstep( coneCosine, penumbraCosine, angleCosine );
}
#if NUM_DIR_LIGHTS > 0
	struct DirectionalLight {
		vec3 direction;
		vec3 color;
	};
	uniform DirectionalLight directionalLights[ NUM_DIR_LIGHTS ];
	void getDirectionalLightInfo( const in DirectionalLight directionalLight, out IncidentLight light ) {
		light.color = directionalLight.color;
		light.direction = directionalLight.direction;
		light.visible = true;
	}
#endif
#if NUM_POINT_LIGHTS > 0
	struct PointLight {
		vec3 position;
		vec3 color;
		float distance;
		float decay;
	};
	uniform PointLight pointLights[ NUM_POINT_LIGHTS ];
	void getPointLightInfo( const in PointLight pointLight, const in vec3 geometryPosition, out IncidentLight light ) {
		vec3 lVector = pointLight.position - geometryPosition;
		light.direction = normalize( lVector );
		float lightDistance = length( lVector );
		light.color = pointLight.color;
		light.color *= getDistanceAttenuation( lightDistance, pointLight.distance, pointLight.decay );
		light.visible = ( light.color != vec3( 0.0 ) );
	}
#endif
#if NUM_SPOT_LIGHTS > 0
	struct SpotLight {
		vec3 position;
		vec3 direction;
		vec3 color;
		float distance;
		float decay;
		float coneCos;
		float penumbraCos;
	};
	uniform SpotLight spotLights[ NUM_SPOT_LIGHTS ];
	void getSpotLightInfo( const in SpotLight spotLight, const in vec3 geometryPosition, out IncidentLight light ) {
		vec3 lVector = spotLight.position - geometryPosition;
		light.direction = normalize( lVector );
		float angleCos = dot( light.direction, spotLight.direction );
		float spotAttenuation = getSpotAttenuation( spotLight.coneCos, spotLight.penumbraCos, angleCos );
		if ( spotAttenuation > 0.0 ) {
			float lightDistance = length( lVector );
			light.color = spotLight.color * spotAttenuation;
			light.color *= getDistanceAttenuation( lightDistance, spotLight.distance, spotLight.decay );
			light.visible = ( light.color != vec3( 0.0 ) );
		} else {
			light.color = vec3( 0.0 );
			light.visible = false;
		}
	}
#endif
#if NUM_RECT_AREA_LIGHTS > 0
	struct RectAreaLight {
		vec3 color;
		vec3 position;
		vec3 halfWidth;
		vec3 halfHeight;
	};
	uniform sampler2D ltc_1;	uniform sampler2D ltc_2;
	uniform RectAreaLight rectAreaLights[ NUM_RECT_AREA_LIGHTS ];
#endif
#if NUM_HEMI_LIGHTS > 0
	struct HemisphereLight {
		vec3 direction;
		vec3 skyColor;
		vec3 groundColor;
	};
	uniform HemisphereLight hemisphereLights[ NUM_HEMI_LIGHTS ];
	vec3 getHemisphereLightIrradiance( const in HemisphereLight hemiLight, const in vec3 normal ) {
		float dotNL = dot( normal, hemiLight.direction );
		float hemiDiffuseWeight = 0.5 * dotNL + 0.5;
		vec3 irradiance = mix( hemiLight.groundColor, hemiLight.skyColor, hemiDiffuseWeight );
		return irradiance;
	}
#endif`,wS=`#ifdef USE_ENVMAP
	vec3 getIBLIrradiance( const in vec3 normal ) {
		#ifdef ENVMAP_TYPE_CUBE_UV
			vec3 worldNormal = inverseTransformDirection( normal, viewMatrix );
			vec4 envMapColor = textureCubeUV( envMap, envMapRotation * worldNormal, 1.0 );
			return PI * envMapColor.rgb * envMapIntensity;
		#else
			return vec3( 0.0 );
		#endif
	}
	vec3 getIBLRadiance( const in vec3 viewDir, const in vec3 normal, const in float roughness ) {
		#ifdef ENVMAP_TYPE_CUBE_UV
			vec3 reflectVec = reflect( - viewDir, normal );
			reflectVec = normalize( mix( reflectVec, normal, roughness * roughness) );
			reflectVec = inverseTransformDirection( reflectVec, viewMatrix );
			vec4 envMapColor = textureCubeUV( envMap, envMapRotation * reflectVec, roughness );
			return envMapColor.rgb * envMapIntensity;
		#else
			return vec3( 0.0 );
		#endif
	}
	#ifdef USE_ANISOTROPY
		vec3 getIBLAnisotropyRadiance( const in vec3 viewDir, const in vec3 normal, const in float roughness, const in vec3 bitangent, const in float anisotropy ) {
			#ifdef ENVMAP_TYPE_CUBE_UV
				vec3 bentNormal = cross( bitangent, viewDir );
				bentNormal = normalize( cross( bentNormal, bitangent ) );
				bentNormal = normalize( mix( bentNormal, normal, pow2( pow2( 1.0 - anisotropy * ( 1.0 - roughness ) ) ) ) );
				return getIBLRadiance( viewDir, bentNormal, roughness );
			#else
				return vec3( 0.0 );
			#endif
		}
	#endif
#endif`,AS=`ToonMaterial material;
material.diffuseColor = diffuseColor.rgb;`,CS=`varying vec3 vViewPosition;
struct ToonMaterial {
	vec3 diffuseColor;
};
void RE_Direct_Toon( const in IncidentLight directLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in ToonMaterial material, inout ReflectedLight reflectedLight ) {
	vec3 irradiance = getGradientIrradiance( geometryNormal, directLight.direction ) * directLight.color;
	reflectedLight.directDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
void RE_IndirectDiffuse_Toon( const in vec3 irradiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in ToonMaterial material, inout ReflectedLight reflectedLight ) {
	reflectedLight.indirectDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
#define RE_Direct				RE_Direct_Toon
#define RE_IndirectDiffuse		RE_IndirectDiffuse_Toon`,RS=`BlinnPhongMaterial material;
material.diffuseColor = diffuseColor.rgb;
material.specularColor = specular;
material.specularShininess = shininess;
material.specularStrength = specularStrength;`,bS=`varying vec3 vViewPosition;
struct BlinnPhongMaterial {
	vec3 diffuseColor;
	vec3 specularColor;
	float specularShininess;
	float specularStrength;
};
void RE_Direct_BlinnPhong( const in IncidentLight directLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in BlinnPhongMaterial material, inout ReflectedLight reflectedLight ) {
	float dotNL = saturate( dot( geometryNormal, directLight.direction ) );
	vec3 irradiance = dotNL * directLight.color;
	reflectedLight.directDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
	reflectedLight.directSpecular += irradiance * BRDF_BlinnPhong( directLight.direction, geometryViewDir, geometryNormal, material.specularColor, material.specularShininess ) * material.specularStrength;
}
void RE_IndirectDiffuse_BlinnPhong( const in vec3 irradiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in BlinnPhongMaterial material, inout ReflectedLight reflectedLight ) {
	reflectedLight.indirectDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
#define RE_Direct				RE_Direct_BlinnPhong
#define RE_IndirectDiffuse		RE_IndirectDiffuse_BlinnPhong`,PS=`PhysicalMaterial material;
material.diffuseColor = diffuseColor.rgb * ( 1.0 - metalnessFactor );
vec3 dxy = max( abs( dFdx( nonPerturbedNormal ) ), abs( dFdy( nonPerturbedNormal ) ) );
float geometryRoughness = max( max( dxy.x, dxy.y ), dxy.z );
material.roughness = max( roughnessFactor, 0.0525 );material.roughness += geometryRoughness;
material.roughness = min( material.roughness, 1.0 );
#ifdef IOR
	material.ior = ior;
	#ifdef USE_SPECULAR
		float specularIntensityFactor = specularIntensity;
		vec3 specularColorFactor = specularColor;
		#ifdef USE_SPECULAR_COLORMAP
			specularColorFactor *= texture2D( specularColorMap, vSpecularColorMapUv ).rgb;
		#endif
		#ifdef USE_SPECULAR_INTENSITYMAP
			specularIntensityFactor *= texture2D( specularIntensityMap, vSpecularIntensityMapUv ).a;
		#endif
		material.specularF90 = mix( specularIntensityFactor, 1.0, metalnessFactor );
	#else
		float specularIntensityFactor = 1.0;
		vec3 specularColorFactor = vec3( 1.0 );
		material.specularF90 = 1.0;
	#endif
	material.specularColor = mix( min( pow2( ( material.ior - 1.0 ) / ( material.ior + 1.0 ) ) * specularColorFactor, vec3( 1.0 ) ) * specularIntensityFactor, diffuseColor.rgb, metalnessFactor );
#else
	material.specularColor = mix( vec3( 0.04 ), diffuseColor.rgb, metalnessFactor );
	material.specularF90 = 1.0;
#endif
#ifdef USE_CLEARCOAT
	material.clearcoat = clearcoat;
	material.clearcoatRoughness = clearcoatRoughness;
	material.clearcoatF0 = vec3( 0.04 );
	material.clearcoatF90 = 1.0;
	#ifdef USE_CLEARCOATMAP
		material.clearcoat *= texture2D( clearcoatMap, vClearcoatMapUv ).x;
	#endif
	#ifdef USE_CLEARCOAT_ROUGHNESSMAP
		material.clearcoatRoughness *= texture2D( clearcoatRoughnessMap, vClearcoatRoughnessMapUv ).y;
	#endif
	material.clearcoat = saturate( material.clearcoat );	material.clearcoatRoughness = max( material.clearcoatRoughness, 0.0525 );
	material.clearcoatRoughness += geometryRoughness;
	material.clearcoatRoughness = min( material.clearcoatRoughness, 1.0 );
#endif
#ifdef USE_DISPERSION
	material.dispersion = dispersion;
#endif
#ifdef USE_IRIDESCENCE
	material.iridescence = iridescence;
	material.iridescenceIOR = iridescenceIOR;
	#ifdef USE_IRIDESCENCEMAP
		material.iridescence *= texture2D( iridescenceMap, vIridescenceMapUv ).r;
	#endif
	#ifdef USE_IRIDESCENCE_THICKNESSMAP
		material.iridescenceThickness = (iridescenceThicknessMaximum - iridescenceThicknessMinimum) * texture2D( iridescenceThicknessMap, vIridescenceThicknessMapUv ).g + iridescenceThicknessMinimum;
	#else
		material.iridescenceThickness = iridescenceThicknessMaximum;
	#endif
#endif
#ifdef USE_SHEEN
	material.sheenColor = sheenColor;
	#ifdef USE_SHEEN_COLORMAP
		material.sheenColor *= texture2D( sheenColorMap, vSheenColorMapUv ).rgb;
	#endif
	material.sheenRoughness = clamp( sheenRoughness, 0.07, 1.0 );
	#ifdef USE_SHEEN_ROUGHNESSMAP
		material.sheenRoughness *= texture2D( sheenRoughnessMap, vSheenRoughnessMapUv ).a;
	#endif
#endif
#ifdef USE_ANISOTROPY
	#ifdef USE_ANISOTROPYMAP
		mat2 anisotropyMat = mat2( anisotropyVector.x, anisotropyVector.y, - anisotropyVector.y, anisotropyVector.x );
		vec3 anisotropyPolar = texture2D( anisotropyMap, vAnisotropyMapUv ).rgb;
		vec2 anisotropyV = anisotropyMat * normalize( 2.0 * anisotropyPolar.rg - vec2( 1.0 ) ) * anisotropyPolar.b;
	#else
		vec2 anisotropyV = anisotropyVector;
	#endif
	material.anisotropy = length( anisotropyV );
	if( material.anisotropy == 0.0 ) {
		anisotropyV = vec2( 1.0, 0.0 );
	} else {
		anisotropyV /= material.anisotropy;
		material.anisotropy = saturate( material.anisotropy );
	}
	material.alphaT = mix( pow2( material.roughness ), 1.0, pow2( material.anisotropy ) );
	material.anisotropyT = tbn[ 0 ] * anisotropyV.x + tbn[ 1 ] * anisotropyV.y;
	material.anisotropyB = tbn[ 1 ] * anisotropyV.x - tbn[ 0 ] * anisotropyV.y;
#endif`,LS=`struct PhysicalMaterial {
	vec3 diffuseColor;
	float roughness;
	vec3 specularColor;
	float specularF90;
	float dispersion;
	#ifdef USE_CLEARCOAT
		float clearcoat;
		float clearcoatRoughness;
		vec3 clearcoatF0;
		float clearcoatF90;
	#endif
	#ifdef USE_IRIDESCENCE
		float iridescence;
		float iridescenceIOR;
		float iridescenceThickness;
		vec3 iridescenceFresnel;
		vec3 iridescenceF0;
	#endif
	#ifdef USE_SHEEN
		vec3 sheenColor;
		float sheenRoughness;
	#endif
	#ifdef IOR
		float ior;
	#endif
	#ifdef USE_TRANSMISSION
		float transmission;
		float transmissionAlpha;
		float thickness;
		float attenuationDistance;
		vec3 attenuationColor;
	#endif
	#ifdef USE_ANISOTROPY
		float anisotropy;
		float alphaT;
		vec3 anisotropyT;
		vec3 anisotropyB;
	#endif
};
vec3 clearcoatSpecularDirect = vec3( 0.0 );
vec3 clearcoatSpecularIndirect = vec3( 0.0 );
vec3 sheenSpecularDirect = vec3( 0.0 );
vec3 sheenSpecularIndirect = vec3(0.0 );
vec3 Schlick_to_F0( const in vec3 f, const in float f90, const in float dotVH ) {
    float x = clamp( 1.0 - dotVH, 0.0, 1.0 );
    float x2 = x * x;
    float x5 = clamp( x * x2 * x2, 0.0, 0.9999 );
    return ( f - vec3( f90 ) * x5 ) / ( 1.0 - x5 );
}
float V_GGX_SmithCorrelated( const in float alpha, const in float dotNL, const in float dotNV ) {
	float a2 = pow2( alpha );
	float gv = dotNL * sqrt( a2 + ( 1.0 - a2 ) * pow2( dotNV ) );
	float gl = dotNV * sqrt( a2 + ( 1.0 - a2 ) * pow2( dotNL ) );
	return 0.5 / max( gv + gl, EPSILON );
}
float D_GGX( const in float alpha, const in float dotNH ) {
	float a2 = pow2( alpha );
	float denom = pow2( dotNH ) * ( a2 - 1.0 ) + 1.0;
	return RECIPROCAL_PI * a2 / pow2( denom );
}
#ifdef USE_ANISOTROPY
	float V_GGX_SmithCorrelated_Anisotropic( const in float alphaT, const in float alphaB, const in float dotTV, const in float dotBV, const in float dotTL, const in float dotBL, const in float dotNV, const in float dotNL ) {
		float gv = dotNL * length( vec3( alphaT * dotTV, alphaB * dotBV, dotNV ) );
		float gl = dotNV * length( vec3( alphaT * dotTL, alphaB * dotBL, dotNL ) );
		float v = 0.5 / ( gv + gl );
		return saturate(v);
	}
	float D_GGX_Anisotropic( const in float alphaT, const in float alphaB, const in float dotNH, const in float dotTH, const in float dotBH ) {
		float a2 = alphaT * alphaB;
		highp vec3 v = vec3( alphaB * dotTH, alphaT * dotBH, a2 * dotNH );
		highp float v2 = dot( v, v );
		float w2 = a2 / v2;
		return RECIPROCAL_PI * a2 * pow2 ( w2 );
	}
#endif
#ifdef USE_CLEARCOAT
	vec3 BRDF_GGX_Clearcoat( const in vec3 lightDir, const in vec3 viewDir, const in vec3 normal, const in PhysicalMaterial material) {
		vec3 f0 = material.clearcoatF0;
		float f90 = material.clearcoatF90;
		float roughness = material.clearcoatRoughness;
		float alpha = pow2( roughness );
		vec3 halfDir = normalize( lightDir + viewDir );
		float dotNL = saturate( dot( normal, lightDir ) );
		float dotNV = saturate( dot( normal, viewDir ) );
		float dotNH = saturate( dot( normal, halfDir ) );
		float dotVH = saturate( dot( viewDir, halfDir ) );
		vec3 F = F_Schlick( f0, f90, dotVH );
		float V = V_GGX_SmithCorrelated( alpha, dotNL, dotNV );
		float D = D_GGX( alpha, dotNH );
		return F * ( V * D );
	}
#endif
vec3 BRDF_GGX( const in vec3 lightDir, const in vec3 viewDir, const in vec3 normal, const in PhysicalMaterial material ) {
	vec3 f0 = material.specularColor;
	float f90 = material.specularF90;
	float roughness = material.roughness;
	float alpha = pow2( roughness );
	vec3 halfDir = normalize( lightDir + viewDir );
	float dotNL = saturate( dot( normal, lightDir ) );
	float dotNV = saturate( dot( normal, viewDir ) );
	float dotNH = saturate( dot( normal, halfDir ) );
	float dotVH = saturate( dot( viewDir, halfDir ) );
	vec3 F = F_Schlick( f0, f90, dotVH );
	#ifdef USE_IRIDESCENCE
		F = mix( F, material.iridescenceFresnel, material.iridescence );
	#endif
	#ifdef USE_ANISOTROPY
		float dotTL = dot( material.anisotropyT, lightDir );
		float dotTV = dot( material.anisotropyT, viewDir );
		float dotTH = dot( material.anisotropyT, halfDir );
		float dotBL = dot( material.anisotropyB, lightDir );
		float dotBV = dot( material.anisotropyB, viewDir );
		float dotBH = dot( material.anisotropyB, halfDir );
		float V = V_GGX_SmithCorrelated_Anisotropic( material.alphaT, alpha, dotTV, dotBV, dotTL, dotBL, dotNV, dotNL );
		float D = D_GGX_Anisotropic( material.alphaT, alpha, dotNH, dotTH, dotBH );
	#else
		float V = V_GGX_SmithCorrelated( alpha, dotNL, dotNV );
		float D = D_GGX( alpha, dotNH );
	#endif
	return F * ( V * D );
}
vec2 LTC_Uv( const in vec3 N, const in vec3 V, const in float roughness ) {
	const float LUT_SIZE = 64.0;
	const float LUT_SCALE = ( LUT_SIZE - 1.0 ) / LUT_SIZE;
	const float LUT_BIAS = 0.5 / LUT_SIZE;
	float dotNV = saturate( dot( N, V ) );
	vec2 uv = vec2( roughness, sqrt( 1.0 - dotNV ) );
	uv = uv * LUT_SCALE + LUT_BIAS;
	return uv;
}
float LTC_ClippedSphereFormFactor( const in vec3 f ) {
	float l = length( f );
	return max( ( l * l + f.z ) / ( l + 1.0 ), 0.0 );
}
vec3 LTC_EdgeVectorFormFactor( const in vec3 v1, const in vec3 v2 ) {
	float x = dot( v1, v2 );
	float y = abs( x );
	float a = 0.8543985 + ( 0.4965155 + 0.0145206 * y ) * y;
	float b = 3.4175940 + ( 4.1616724 + y ) * y;
	float v = a / b;
	float theta_sintheta = ( x > 0.0 ) ? v : 0.5 * inversesqrt( max( 1.0 - x * x, 1e-7 ) ) - v;
	return cross( v1, v2 ) * theta_sintheta;
}
vec3 LTC_Evaluate( const in vec3 N, const in vec3 V, const in vec3 P, const in mat3 mInv, const in vec3 rectCoords[ 4 ] ) {
	vec3 v1 = rectCoords[ 1 ] - rectCoords[ 0 ];
	vec3 v2 = rectCoords[ 3 ] - rectCoords[ 0 ];
	vec3 lightNormal = cross( v1, v2 );
	if( dot( lightNormal, P - rectCoords[ 0 ] ) < 0.0 ) return vec3( 0.0 );
	vec3 T1, T2;
	T1 = normalize( V - N * dot( V, N ) );
	T2 = - cross( N, T1 );
	mat3 mat = mInv * transposeMat3( mat3( T1, T2, N ) );
	vec3 coords[ 4 ];
	coords[ 0 ] = mat * ( rectCoords[ 0 ] - P );
	coords[ 1 ] = mat * ( rectCoords[ 1 ] - P );
	coords[ 2 ] = mat * ( rectCoords[ 2 ] - P );
	coords[ 3 ] = mat * ( rectCoords[ 3 ] - P );
	coords[ 0 ] = normalize( coords[ 0 ] );
	coords[ 1 ] = normalize( coords[ 1 ] );
	coords[ 2 ] = normalize( coords[ 2 ] );
	coords[ 3 ] = normalize( coords[ 3 ] );
	vec3 vectorFormFactor = vec3( 0.0 );
	vectorFormFactor += LTC_EdgeVectorFormFactor( coords[ 0 ], coords[ 1 ] );
	vectorFormFactor += LTC_EdgeVectorFormFactor( coords[ 1 ], coords[ 2 ] );
	vectorFormFactor += LTC_EdgeVectorFormFactor( coords[ 2 ], coords[ 3 ] );
	vectorFormFactor += LTC_EdgeVectorFormFactor( coords[ 3 ], coords[ 0 ] );
	float result = LTC_ClippedSphereFormFactor( vectorFormFactor );
	return vec3( result );
}
#if defined( USE_SHEEN )
float D_Charlie( float roughness, float dotNH ) {
	float alpha = pow2( roughness );
	float invAlpha = 1.0 / alpha;
	float cos2h = dotNH * dotNH;
	float sin2h = max( 1.0 - cos2h, 0.0078125 );
	return ( 2.0 + invAlpha ) * pow( sin2h, invAlpha * 0.5 ) / ( 2.0 * PI );
}
float V_Neubelt( float dotNV, float dotNL ) {
	return saturate( 1.0 / ( 4.0 * ( dotNL + dotNV - dotNL * dotNV ) ) );
}
vec3 BRDF_Sheen( const in vec3 lightDir, const in vec3 viewDir, const in vec3 normal, vec3 sheenColor, const in float sheenRoughness ) {
	vec3 halfDir = normalize( lightDir + viewDir );
	float dotNL = saturate( dot( normal, lightDir ) );
	float dotNV = saturate( dot( normal, viewDir ) );
	float dotNH = saturate( dot( normal, halfDir ) );
	float D = D_Charlie( sheenRoughness, dotNH );
	float V = V_Neubelt( dotNV, dotNL );
	return sheenColor * ( D * V );
}
#endif
float IBLSheenBRDF( const in vec3 normal, const in vec3 viewDir, const in float roughness ) {
	float dotNV = saturate( dot( normal, viewDir ) );
	float r2 = roughness * roughness;
	float a = roughness < 0.25 ? -339.2 * r2 + 161.4 * roughness - 25.9 : -8.48 * r2 + 14.3 * roughness - 9.95;
	float b = roughness < 0.25 ? 44.0 * r2 - 23.7 * roughness + 3.26 : 1.97 * r2 - 3.27 * roughness + 0.72;
	float DG = exp( a * dotNV + b ) + ( roughness < 0.25 ? 0.0 : 0.1 * ( roughness - 0.25 ) );
	return saturate( DG * RECIPROCAL_PI );
}
vec2 DFGApprox( const in vec3 normal, const in vec3 viewDir, const in float roughness ) {
	float dotNV = saturate( dot( normal, viewDir ) );
	const vec4 c0 = vec4( - 1, - 0.0275, - 0.572, 0.022 );
	const vec4 c1 = vec4( 1, 0.0425, 1.04, - 0.04 );
	vec4 r = roughness * c0 + c1;
	float a004 = min( r.x * r.x, exp2( - 9.28 * dotNV ) ) * r.x + r.y;
	vec2 fab = vec2( - 1.04, 1.04 ) * a004 + r.zw;
	return fab;
}
vec3 EnvironmentBRDF( const in vec3 normal, const in vec3 viewDir, const in vec3 specularColor, const in float specularF90, const in float roughness ) {
	vec2 fab = DFGApprox( normal, viewDir, roughness );
	return specularColor * fab.x + specularF90 * fab.y;
}
#ifdef USE_IRIDESCENCE
void computeMultiscatteringIridescence( const in vec3 normal, const in vec3 viewDir, const in vec3 specularColor, const in float specularF90, const in float iridescence, const in vec3 iridescenceF0, const in float roughness, inout vec3 singleScatter, inout vec3 multiScatter ) {
#else
void computeMultiscattering( const in vec3 normal, const in vec3 viewDir, const in vec3 specularColor, const in float specularF90, const in float roughness, inout vec3 singleScatter, inout vec3 multiScatter ) {
#endif
	vec2 fab = DFGApprox( normal, viewDir, roughness );
	#ifdef USE_IRIDESCENCE
		vec3 Fr = mix( specularColor, iridescenceF0, iridescence );
	#else
		vec3 Fr = specularColor;
	#endif
	vec3 FssEss = Fr * fab.x + specularF90 * fab.y;
	float Ess = fab.x + fab.y;
	float Ems = 1.0 - Ess;
	vec3 Favg = Fr + ( 1.0 - Fr ) * 0.047619;	vec3 Fms = FssEss * Favg / ( 1.0 - Ems * Favg );
	singleScatter += FssEss;
	multiScatter += Fms * Ems;
}
#if NUM_RECT_AREA_LIGHTS > 0
	void RE_Direct_RectArea_Physical( const in RectAreaLight rectAreaLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in PhysicalMaterial material, inout ReflectedLight reflectedLight ) {
		vec3 normal = geometryNormal;
		vec3 viewDir = geometryViewDir;
		vec3 position = geometryPosition;
		vec3 lightPos = rectAreaLight.position;
		vec3 halfWidth = rectAreaLight.halfWidth;
		vec3 halfHeight = rectAreaLight.halfHeight;
		vec3 lightColor = rectAreaLight.color;
		float roughness = material.roughness;
		vec3 rectCoords[ 4 ];
		rectCoords[ 0 ] = lightPos + halfWidth - halfHeight;		rectCoords[ 1 ] = lightPos - halfWidth - halfHeight;
		rectCoords[ 2 ] = lightPos - halfWidth + halfHeight;
		rectCoords[ 3 ] = lightPos + halfWidth + halfHeight;
		vec2 uv = LTC_Uv( normal, viewDir, roughness );
		vec4 t1 = texture2D( ltc_1, uv );
		vec4 t2 = texture2D( ltc_2, uv );
		mat3 mInv = mat3(
			vec3( t1.x, 0, t1.y ),
			vec3(    0, 1,    0 ),
			vec3( t1.z, 0, t1.w )
		);
		vec3 fresnel = ( material.specularColor * t2.x + ( vec3( 1.0 ) - material.specularColor ) * t2.y );
		reflectedLight.directSpecular += lightColor * fresnel * LTC_Evaluate( normal, viewDir, position, mInv, rectCoords );
		reflectedLight.directDiffuse += lightColor * material.diffuseColor * LTC_Evaluate( normal, viewDir, position, mat3( 1.0 ), rectCoords );
	}
#endif
void RE_Direct_Physical( const in IncidentLight directLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in PhysicalMaterial material, inout ReflectedLight reflectedLight ) {
	float dotNL = saturate( dot( geometryNormal, directLight.direction ) );
	vec3 irradiance = dotNL * directLight.color;
	#ifdef USE_CLEARCOAT
		float dotNLcc = saturate( dot( geometryClearcoatNormal, directLight.direction ) );
		vec3 ccIrradiance = dotNLcc * directLight.color;
		clearcoatSpecularDirect += ccIrradiance * BRDF_GGX_Clearcoat( directLight.direction, geometryViewDir, geometryClearcoatNormal, material );
	#endif
	#ifdef USE_SHEEN
		sheenSpecularDirect += irradiance * BRDF_Sheen( directLight.direction, geometryViewDir, geometryNormal, material.sheenColor, material.sheenRoughness );
	#endif
	reflectedLight.directSpecular += irradiance * BRDF_GGX( directLight.direction, geometryViewDir, geometryNormal, material );
	reflectedLight.directDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
void RE_IndirectDiffuse_Physical( const in vec3 irradiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in PhysicalMaterial material, inout ReflectedLight reflectedLight ) {
	reflectedLight.indirectDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
void RE_IndirectSpecular_Physical( const in vec3 radiance, const in vec3 irradiance, const in vec3 clearcoatRadiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in PhysicalMaterial material, inout ReflectedLight reflectedLight) {
	#ifdef USE_CLEARCOAT
		clearcoatSpecularIndirect += clearcoatRadiance * EnvironmentBRDF( geometryClearcoatNormal, geometryViewDir, material.clearcoatF0, material.clearcoatF90, material.clearcoatRoughness );
	#endif
	#ifdef USE_SHEEN
		sheenSpecularIndirect += irradiance * material.sheenColor * IBLSheenBRDF( geometryNormal, geometryViewDir, material.sheenRoughness );
	#endif
	vec3 singleScattering = vec3( 0.0 );
	vec3 multiScattering = vec3( 0.0 );
	vec3 cosineWeightedIrradiance = irradiance * RECIPROCAL_PI;
	#ifdef USE_IRIDESCENCE
		computeMultiscatteringIridescence( geometryNormal, geometryViewDir, material.specularColor, material.specularF90, material.iridescence, material.iridescenceFresnel, material.roughness, singleScattering, multiScattering );
	#else
		computeMultiscattering( geometryNormal, geometryViewDir, material.specularColor, material.specularF90, material.roughness, singleScattering, multiScattering );
	#endif
	vec3 totalScattering = singleScattering + multiScattering;
	vec3 diffuse = material.diffuseColor * ( 1.0 - max( max( totalScattering.r, totalScattering.g ), totalScattering.b ) );
	reflectedLight.indirectSpecular += radiance * singleScattering;
	reflectedLight.indirectSpecular += multiScattering * cosineWeightedIrradiance;
	reflectedLight.indirectDiffuse += diffuse * cosineWeightedIrradiance;
}
#define RE_Direct				RE_Direct_Physical
#define RE_Direct_RectArea		RE_Direct_RectArea_Physical
#define RE_IndirectDiffuse		RE_IndirectDiffuse_Physical
#define RE_IndirectSpecular		RE_IndirectSpecular_Physical
float computeSpecularOcclusion( const in float dotNV, const in float ambientOcclusion, const in float roughness ) {
	return saturate( pow( dotNV + ambientOcclusion, exp2( - 16.0 * roughness - 1.0 ) ) - 1.0 + ambientOcclusion );
}`,DS=`
vec3 geometryPosition = - vViewPosition;
vec3 geometryNormal = normal;
vec3 geometryViewDir = ( isOrthographic ) ? vec3( 0, 0, 1 ) : normalize( vViewPosition );
vec3 geometryClearcoatNormal = vec3( 0.0 );
#ifdef USE_CLEARCOAT
	geometryClearcoatNormal = clearcoatNormal;
#endif
#ifdef USE_IRIDESCENCE
	float dotNVi = saturate( dot( normal, geometryViewDir ) );
	if ( material.iridescenceThickness == 0.0 ) {
		material.iridescence = 0.0;
	} else {
		material.iridescence = saturate( material.iridescence );
	}
	if ( material.iridescence > 0.0 ) {
		material.iridescenceFresnel = evalIridescence( 1.0, material.iridescenceIOR, dotNVi, material.iridescenceThickness, material.specularColor );
		material.iridescenceF0 = Schlick_to_F0( material.iridescenceFresnel, 1.0, dotNVi );
	}
#endif
IncidentLight directLight;
#if ( NUM_POINT_LIGHTS > 0 ) && defined( RE_Direct )
	PointLight pointLight;
	#if defined( USE_SHADOWMAP ) && NUM_POINT_LIGHT_SHADOWS > 0
	PointLightShadow pointLightShadow;
	#endif
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_POINT_LIGHTS; i ++ ) {
		pointLight = pointLights[ i ];
		getPointLightInfo( pointLight, geometryPosition, directLight );
		#if defined( USE_SHADOWMAP ) && ( UNROLLED_LOOP_INDEX < NUM_POINT_LIGHT_SHADOWS )
		pointLightShadow = pointLightShadows[ i ];
		directLight.color *= ( directLight.visible && receiveShadow ) ? getPointShadow( pointShadowMap[ i ], pointLightShadow.shadowMapSize, pointLightShadow.shadowIntensity, pointLightShadow.shadowBias, pointLightShadow.shadowRadius, vPointShadowCoord[ i ], pointLightShadow.shadowCameraNear, pointLightShadow.shadowCameraFar ) : 1.0;
		#endif
		RE_Direct( directLight, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
	}
	#pragma unroll_loop_end
#endif
#if ( NUM_SPOT_LIGHTS > 0 ) && defined( RE_Direct )
	SpotLight spotLight;
	vec4 spotColor;
	vec3 spotLightCoord;
	bool inSpotLightMap;
	#if defined( USE_SHADOWMAP ) && NUM_SPOT_LIGHT_SHADOWS > 0
	SpotLightShadow spotLightShadow;
	#endif
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_SPOT_LIGHTS; i ++ ) {
		spotLight = spotLights[ i ];
		getSpotLightInfo( spotLight, geometryPosition, directLight );
		#if ( UNROLLED_LOOP_INDEX < NUM_SPOT_LIGHT_SHADOWS_WITH_MAPS )
		#define SPOT_LIGHT_MAP_INDEX UNROLLED_LOOP_INDEX
		#elif ( UNROLLED_LOOP_INDEX < NUM_SPOT_LIGHT_SHADOWS )
		#define SPOT_LIGHT_MAP_INDEX NUM_SPOT_LIGHT_MAPS
		#else
		#define SPOT_LIGHT_MAP_INDEX ( UNROLLED_LOOP_INDEX - NUM_SPOT_LIGHT_SHADOWS + NUM_SPOT_LIGHT_SHADOWS_WITH_MAPS )
		#endif
		#if ( SPOT_LIGHT_MAP_INDEX < NUM_SPOT_LIGHT_MAPS )
			spotLightCoord = vSpotLightCoord[ i ].xyz / vSpotLightCoord[ i ].w;
			inSpotLightMap = all( lessThan( abs( spotLightCoord * 2. - 1. ), vec3( 1.0 ) ) );
			spotColor = texture2D( spotLightMap[ SPOT_LIGHT_MAP_INDEX ], spotLightCoord.xy );
			directLight.color = inSpotLightMap ? directLight.color * spotColor.rgb : directLight.color;
		#endif
		#undef SPOT_LIGHT_MAP_INDEX
		#if defined( USE_SHADOWMAP ) && ( UNROLLED_LOOP_INDEX < NUM_SPOT_LIGHT_SHADOWS )
		spotLightShadow = spotLightShadows[ i ];
		directLight.color *= ( directLight.visible && receiveShadow ) ? getShadow( spotShadowMap[ i ], spotLightShadow.shadowMapSize, spotLightShadow.shadowIntensity, spotLightShadow.shadowBias, spotLightShadow.shadowRadius, vSpotLightCoord[ i ] ) : 1.0;
		#endif
		RE_Direct( directLight, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
	}
	#pragma unroll_loop_end
#endif
#if ( NUM_DIR_LIGHTS > 0 ) && defined( RE_Direct )
	DirectionalLight directionalLight;
	#if defined( USE_SHADOWMAP ) && NUM_DIR_LIGHT_SHADOWS > 0
	DirectionalLightShadow directionalLightShadow;
	#endif
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_DIR_LIGHTS; i ++ ) {
		directionalLight = directionalLights[ i ];
		getDirectionalLightInfo( directionalLight, directLight );
		#if defined( USE_SHADOWMAP ) && ( UNROLLED_LOOP_INDEX < NUM_DIR_LIGHT_SHADOWS )
		directionalLightShadow = directionalLightShadows[ i ];
		directLight.color *= ( directLight.visible && receiveShadow ) ? getShadow( directionalShadowMap[ i ], directionalLightShadow.shadowMapSize, directionalLightShadow.shadowIntensity, directionalLightShadow.shadowBias, directionalLightShadow.shadowRadius, vDirectionalShadowCoord[ i ] ) : 1.0;
		#endif
		RE_Direct( directLight, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
	}
	#pragma unroll_loop_end
#endif
#if ( NUM_RECT_AREA_LIGHTS > 0 ) && defined( RE_Direct_RectArea )
	RectAreaLight rectAreaLight;
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_RECT_AREA_LIGHTS; i ++ ) {
		rectAreaLight = rectAreaLights[ i ];
		RE_Direct_RectArea( rectAreaLight, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
	}
	#pragma unroll_loop_end
#endif
#if defined( RE_IndirectDiffuse )
	vec3 iblIrradiance = vec3( 0.0 );
	vec3 irradiance = getAmbientLightIrradiance( ambientLightColor );
	#if defined( USE_LIGHT_PROBES )
		irradiance += getLightProbeIrradiance( lightProbe, geometryNormal );
	#endif
	#if ( NUM_HEMI_LIGHTS > 0 )
		#pragma unroll_loop_start
		for ( int i = 0; i < NUM_HEMI_LIGHTS; i ++ ) {
			irradiance += getHemisphereLightIrradiance( hemisphereLights[ i ], geometryNormal );
		}
		#pragma unroll_loop_end
	#endif
#endif
#if defined( RE_IndirectSpecular )
	vec3 radiance = vec3( 0.0 );
	vec3 clearcoatRadiance = vec3( 0.0 );
#endif`,NS=`#if defined( RE_IndirectDiffuse )
	#ifdef USE_LIGHTMAP
		vec4 lightMapTexel = texture2D( lightMap, vLightMapUv );
		vec3 lightMapIrradiance = lightMapTexel.rgb * lightMapIntensity;
		irradiance += lightMapIrradiance;
	#endif
	#if defined( USE_ENVMAP ) && defined( STANDARD ) && defined( ENVMAP_TYPE_CUBE_UV )
		iblIrradiance += getIBLIrradiance( geometryNormal );
	#endif
#endif
#if defined( USE_ENVMAP ) && defined( RE_IndirectSpecular )
	#ifdef USE_ANISOTROPY
		radiance += getIBLAnisotropyRadiance( geometryViewDir, geometryNormal, material.roughness, material.anisotropyB, material.anisotropy );
	#else
		radiance += getIBLRadiance( geometryViewDir, geometryNormal, material.roughness );
	#endif
	#ifdef USE_CLEARCOAT
		clearcoatRadiance += getIBLRadiance( geometryViewDir, geometryClearcoatNormal, material.clearcoatRoughness );
	#endif
#endif`,IS=`#if defined( RE_IndirectDiffuse )
	RE_IndirectDiffuse( irradiance, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
#endif
#if defined( RE_IndirectSpecular )
	RE_IndirectSpecular( radiance, iblIrradiance, clearcoatRadiance, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
#endif`,US=`#if defined( USE_LOGDEPTHBUF )
	gl_FragDepth = vIsPerspective == 0.0 ? gl_FragCoord.z : log2( vFragDepth ) * logDepthBufFC * 0.5;
#endif`,FS=`#if defined( USE_LOGDEPTHBUF )
	uniform float logDepthBufFC;
	varying float vFragDepth;
	varying float vIsPerspective;
#endif`,OS=`#ifdef USE_LOGDEPTHBUF
	varying float vFragDepth;
	varying float vIsPerspective;
#endif`,kS=`#ifdef USE_LOGDEPTHBUF
	vFragDepth = 1.0 + gl_Position.w;
	vIsPerspective = float( isPerspectiveMatrix( projectionMatrix ) );
#endif`,BS=`#ifdef USE_MAP
	vec4 sampledDiffuseColor = texture2D( map, vMapUv );
	#ifdef DECODE_VIDEO_TEXTURE
		sampledDiffuseColor = vec4( mix( pow( sampledDiffuseColor.rgb * 0.9478672986 + vec3( 0.0521327014 ), vec3( 2.4 ) ), sampledDiffuseColor.rgb * 0.0773993808, vec3( lessThanEqual( sampledDiffuseColor.rgb, vec3( 0.04045 ) ) ) ), sampledDiffuseColor.w );
	
	#endif
	diffuseColor *= sampledDiffuseColor;
#endif`,zS=`#ifdef USE_MAP
	uniform sampler2D map;
#endif`,HS=`#if defined( USE_MAP ) || defined( USE_ALPHAMAP )
	#if defined( USE_POINTS_UV )
		vec2 uv = vUv;
	#else
		vec2 uv = ( uvTransform * vec3( gl_PointCoord.x, 1.0 - gl_PointCoord.y, 1 ) ).xy;
	#endif
#endif
#ifdef USE_MAP
	diffuseColor *= texture2D( map, uv );
#endif
#ifdef USE_ALPHAMAP
	diffuseColor.a *= texture2D( alphaMap, uv ).g;
#endif`,VS=`#if defined( USE_POINTS_UV )
	varying vec2 vUv;
#else
	#if defined( USE_MAP ) || defined( USE_ALPHAMAP )
		uniform mat3 uvTransform;
	#endif
#endif
#ifdef USE_MAP
	uniform sampler2D map;
#endif
#ifdef USE_ALPHAMAP
	uniform sampler2D alphaMap;
#endif`,GS=`float metalnessFactor = metalness;
#ifdef USE_METALNESSMAP
	vec4 texelMetalness = texture2D( metalnessMap, vMetalnessMapUv );
	metalnessFactor *= texelMetalness.b;
#endif`,WS=`#ifdef USE_METALNESSMAP
	uniform sampler2D metalnessMap;
#endif`,jS=`#ifdef USE_INSTANCING_MORPH
	float morphTargetInfluences[ MORPHTARGETS_COUNT ];
	float morphTargetBaseInfluence = texelFetch( morphTexture, ivec2( 0, gl_InstanceID ), 0 ).r;
	for ( int i = 0; i < MORPHTARGETS_COUNT; i ++ ) {
		morphTargetInfluences[i] =  texelFetch( morphTexture, ivec2( i + 1, gl_InstanceID ), 0 ).r;
	}
#endif`,XS=`#if defined( USE_MORPHCOLORS )
	vColor *= morphTargetBaseInfluence;
	for ( int i = 0; i < MORPHTARGETS_COUNT; i ++ ) {
		#if defined( USE_COLOR_ALPHA )
			if ( morphTargetInfluences[ i ] != 0.0 ) vColor += getMorph( gl_VertexID, i, 2 ) * morphTargetInfluences[ i ];
		#elif defined( USE_COLOR )
			if ( morphTargetInfluences[ i ] != 0.0 ) vColor += getMorph( gl_VertexID, i, 2 ).rgb * morphTargetInfluences[ i ];
		#endif
	}
#endif`,YS=`#ifdef USE_MORPHNORMALS
	objectNormal *= morphTargetBaseInfluence;
	for ( int i = 0; i < MORPHTARGETS_COUNT; i ++ ) {
		if ( morphTargetInfluences[ i ] != 0.0 ) objectNormal += getMorph( gl_VertexID, i, 1 ).xyz * morphTargetInfluences[ i ];
	}
#endif`,qS=`#ifdef USE_MORPHTARGETS
	#ifndef USE_INSTANCING_MORPH
		uniform float morphTargetBaseInfluence;
		uniform float morphTargetInfluences[ MORPHTARGETS_COUNT ];
	#endif
	uniform sampler2DArray morphTargetsTexture;
	uniform ivec2 morphTargetsTextureSize;
	vec4 getMorph( const in int vertexIndex, const in int morphTargetIndex, const in int offset ) {
		int texelIndex = vertexIndex * MORPHTARGETS_TEXTURE_STRIDE + offset;
		int y = texelIndex / morphTargetsTextureSize.x;
		int x = texelIndex - y * morphTargetsTextureSize.x;
		ivec3 morphUV = ivec3( x, y, morphTargetIndex );
		return texelFetch( morphTargetsTexture, morphUV, 0 );
	}
#endif`,$S=`#ifdef USE_MORPHTARGETS
	transformed *= morphTargetBaseInfluence;
	for ( int i = 0; i < MORPHTARGETS_COUNT; i ++ ) {
		if ( morphTargetInfluences[ i ] != 0.0 ) transformed += getMorph( gl_VertexID, i, 0 ).xyz * morphTargetInfluences[ i ];
	}
#endif`,KS=`float faceDirection = gl_FrontFacing ? 1.0 : - 1.0;
#ifdef FLAT_SHADED
	vec3 fdx = dFdx( vViewPosition );
	vec3 fdy = dFdy( vViewPosition );
	vec3 normal = normalize( cross( fdx, fdy ) );
#else
	vec3 normal = normalize( vNormal );
	#ifdef DOUBLE_SIDED
		normal *= faceDirection;
	#endif
#endif
#if defined( USE_NORMALMAP_TANGENTSPACE ) || defined( USE_CLEARCOAT_NORMALMAP ) || defined( USE_ANISOTROPY )
	#ifdef USE_TANGENT
		mat3 tbn = mat3( normalize( vTangent ), normalize( vBitangent ), normal );
	#else
		mat3 tbn = getTangentFrame( - vViewPosition, normal,
		#if defined( USE_NORMALMAP )
			vNormalMapUv
		#elif defined( USE_CLEARCOAT_NORMALMAP )
			vClearcoatNormalMapUv
		#else
			vUv
		#endif
		);
	#endif
	#if defined( DOUBLE_SIDED ) && ! defined( FLAT_SHADED )
		tbn[0] *= faceDirection;
		tbn[1] *= faceDirection;
	#endif
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	#ifdef USE_TANGENT
		mat3 tbn2 = mat3( normalize( vTangent ), normalize( vBitangent ), normal );
	#else
		mat3 tbn2 = getTangentFrame( - vViewPosition, normal, vClearcoatNormalMapUv );
	#endif
	#if defined( DOUBLE_SIDED ) && ! defined( FLAT_SHADED )
		tbn2[0] *= faceDirection;
		tbn2[1] *= faceDirection;
	#endif
#endif
vec3 nonPerturbedNormal = normal;`,ZS=`#ifdef USE_NORMALMAP_OBJECTSPACE
	normal = texture2D( normalMap, vNormalMapUv ).xyz * 2.0 - 1.0;
	#ifdef FLIP_SIDED
		normal = - normal;
	#endif
	#ifdef DOUBLE_SIDED
		normal = normal * faceDirection;
	#endif
	normal = normalize( normalMatrix * normal );
#elif defined( USE_NORMALMAP_TANGENTSPACE )
	vec3 mapN = texture2D( normalMap, vNormalMapUv ).xyz * 2.0 - 1.0;
	mapN.xy *= normalScale;
	normal = normalize( tbn * mapN );
#elif defined( USE_BUMPMAP )
	normal = perturbNormalArb( - vViewPosition, normal, dHdxy_fwd(), faceDirection );
#endif`,QS=`#ifndef FLAT_SHADED
	varying vec3 vNormal;
	#ifdef USE_TANGENT
		varying vec3 vTangent;
		varying vec3 vBitangent;
	#endif
#endif`,JS=`#ifndef FLAT_SHADED
	varying vec3 vNormal;
	#ifdef USE_TANGENT
		varying vec3 vTangent;
		varying vec3 vBitangent;
	#endif
#endif`,eM=`#ifndef FLAT_SHADED
	vNormal = normalize( transformedNormal );
	#ifdef USE_TANGENT
		vTangent = normalize( transformedTangent );
		vBitangent = normalize( cross( vNormal, vTangent ) * tangent.w );
	#endif
#endif`,tM=`#ifdef USE_NORMALMAP
	uniform sampler2D normalMap;
	uniform vec2 normalScale;
#endif
#ifdef USE_NORMALMAP_OBJECTSPACE
	uniform mat3 normalMatrix;
#endif
#if ! defined ( USE_TANGENT ) && ( defined ( USE_NORMALMAP_TANGENTSPACE ) || defined ( USE_CLEARCOAT_NORMALMAP ) || defined( USE_ANISOTROPY ) )
	mat3 getTangentFrame( vec3 eye_pos, vec3 surf_norm, vec2 uv ) {
		vec3 q0 = dFdx( eye_pos.xyz );
		vec3 q1 = dFdy( eye_pos.xyz );
		vec2 st0 = dFdx( uv.st );
		vec2 st1 = dFdy( uv.st );
		vec3 N = surf_norm;
		vec3 q1perp = cross( q1, N );
		vec3 q0perp = cross( N, q0 );
		vec3 T = q1perp * st0.x + q0perp * st1.x;
		vec3 B = q1perp * st0.y + q0perp * st1.y;
		float det = max( dot( T, T ), dot( B, B ) );
		float scale = ( det == 0.0 ) ? 0.0 : inversesqrt( det );
		return mat3( T * scale, B * scale, N );
	}
#endif`,nM=`#ifdef USE_CLEARCOAT
	vec3 clearcoatNormal = nonPerturbedNormal;
#endif`,iM=`#ifdef USE_CLEARCOAT_NORMALMAP
	vec3 clearcoatMapN = texture2D( clearcoatNormalMap, vClearcoatNormalMapUv ).xyz * 2.0 - 1.0;
	clearcoatMapN.xy *= clearcoatNormalScale;
	clearcoatNormal = normalize( tbn2 * clearcoatMapN );
#endif`,rM=`#ifdef USE_CLEARCOATMAP
	uniform sampler2D clearcoatMap;
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	uniform sampler2D clearcoatNormalMap;
	uniform vec2 clearcoatNormalScale;
#endif
#ifdef USE_CLEARCOAT_ROUGHNESSMAP
	uniform sampler2D clearcoatRoughnessMap;
#endif`,sM=`#ifdef USE_IRIDESCENCEMAP
	uniform sampler2D iridescenceMap;
#endif
#ifdef USE_IRIDESCENCE_THICKNESSMAP
	uniform sampler2D iridescenceThicknessMap;
#endif`,oM=`#ifdef OPAQUE
diffuseColor.a = 1.0;
#endif
#ifdef USE_TRANSMISSION
diffuseColor.a *= material.transmissionAlpha;
#endif
gl_FragColor = vec4( outgoingLight, diffuseColor.a );`,aM=`vec3 packNormalToRGB( const in vec3 normal ) {
	return normalize( normal ) * 0.5 + 0.5;
}
vec3 unpackRGBToNormal( const in vec3 rgb ) {
	return 2.0 * rgb.xyz - 1.0;
}
const float PackUpscale = 256. / 255.;const float UnpackDownscale = 255. / 256.;const float ShiftRight8 = 1. / 256.;
const float Inv255 = 1. / 255.;
const vec4 PackFactors = vec4( 1.0, 256.0, 256.0 * 256.0, 256.0 * 256.0 * 256.0 );
const vec2 UnpackFactors2 = vec2( UnpackDownscale, 1.0 / PackFactors.g );
const vec3 UnpackFactors3 = vec3( UnpackDownscale / PackFactors.rg, 1.0 / PackFactors.b );
const vec4 UnpackFactors4 = vec4( UnpackDownscale / PackFactors.rgb, 1.0 / PackFactors.a );
vec4 packDepthToRGBA( const in float v ) {
	if( v <= 0.0 )
		return vec4( 0., 0., 0., 0. );
	if( v >= 1.0 )
		return vec4( 1., 1., 1., 1. );
	float vuf;
	float af = modf( v * PackFactors.a, vuf );
	float bf = modf( vuf * ShiftRight8, vuf );
	float gf = modf( vuf * ShiftRight8, vuf );
	return vec4( vuf * Inv255, gf * PackUpscale, bf * PackUpscale, af );
}
vec3 packDepthToRGB( const in float v ) {
	if( v <= 0.0 )
		return vec3( 0., 0., 0. );
	if( v >= 1.0 )
		return vec3( 1., 1., 1. );
	float vuf;
	float bf = modf( v * PackFactors.b, vuf );
	float gf = modf( vuf * ShiftRight8, vuf );
	return vec3( vuf * Inv255, gf * PackUpscale, bf );
}
vec2 packDepthToRG( const in float v ) {
	if( v <= 0.0 )
		return vec2( 0., 0. );
	if( v >= 1.0 )
		return vec2( 1., 1. );
	float vuf;
	float gf = modf( v * 256., vuf );
	return vec2( vuf * Inv255, gf );
}
float unpackRGBAToDepth( const in vec4 v ) {
	return dot( v, UnpackFactors4 );
}
float unpackRGBToDepth( const in vec3 v ) {
	return dot( v, UnpackFactors3 );
}
float unpackRGToDepth( const in vec2 v ) {
	return v.r * UnpackFactors2.r + v.g * UnpackFactors2.g;
}
vec4 pack2HalfToRGBA( const in vec2 v ) {
	vec4 r = vec4( v.x, fract( v.x * 255.0 ), v.y, fract( v.y * 255.0 ) );
	return vec4( r.x - r.y / 255.0, r.y, r.z - r.w / 255.0, r.w );
}
vec2 unpackRGBATo2Half( const in vec4 v ) {
	return vec2( v.x + ( v.y / 255.0 ), v.z + ( v.w / 255.0 ) );
}
float viewZToOrthographicDepth( const in float viewZ, const in float near, const in float far ) {
	return ( viewZ + near ) / ( near - far );
}
float orthographicDepthToViewZ( const in float depth, const in float near, const in float far ) {
	return depth * ( near - far ) - near;
}
float viewZToPerspectiveDepth( const in float viewZ, const in float near, const in float far ) {
	return ( ( near + viewZ ) * far ) / ( ( far - near ) * viewZ );
}
float perspectiveDepthToViewZ( const in float depth, const in float near, const in float far ) {
	return ( near * far ) / ( ( far - near ) * depth - far );
}`,lM=`#ifdef PREMULTIPLIED_ALPHA
	gl_FragColor.rgb *= gl_FragColor.a;
#endif`,cM=`vec4 mvPosition = vec4( transformed, 1.0 );
#ifdef USE_BATCHING
	mvPosition = batchingMatrix * mvPosition;
#endif
#ifdef USE_INSTANCING
	mvPosition = instanceMatrix * mvPosition;
#endif
mvPosition = modelViewMatrix * mvPosition;
gl_Position = projectionMatrix * mvPosition;`,uM=`#ifdef DITHERING
	gl_FragColor.rgb = dithering( gl_FragColor.rgb );
#endif`,fM=`#ifdef DITHERING
	vec3 dithering( vec3 color ) {
		float grid_position = rand( gl_FragCoord.xy );
		vec3 dither_shift_RGB = vec3( 0.25 / 255.0, -0.25 / 255.0, 0.25 / 255.0 );
		dither_shift_RGB = mix( 2.0 * dither_shift_RGB, -2.0 * dither_shift_RGB, grid_position );
		return color + dither_shift_RGB;
	}
#endif`,dM=`float roughnessFactor = roughness;
#ifdef USE_ROUGHNESSMAP
	vec4 texelRoughness = texture2D( roughnessMap, vRoughnessMapUv );
	roughnessFactor *= texelRoughness.g;
#endif`,hM=`#ifdef USE_ROUGHNESSMAP
	uniform sampler2D roughnessMap;
#endif`,pM=`#if NUM_SPOT_LIGHT_COORDS > 0
	varying vec4 vSpotLightCoord[ NUM_SPOT_LIGHT_COORDS ];
#endif
#if NUM_SPOT_LIGHT_MAPS > 0
	uniform sampler2D spotLightMap[ NUM_SPOT_LIGHT_MAPS ];
#endif
#ifdef USE_SHADOWMAP
	#if NUM_DIR_LIGHT_SHADOWS > 0
		uniform sampler2D directionalShadowMap[ NUM_DIR_LIGHT_SHADOWS ];
		varying vec4 vDirectionalShadowCoord[ NUM_DIR_LIGHT_SHADOWS ];
		struct DirectionalLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
		};
		uniform DirectionalLightShadow directionalLightShadows[ NUM_DIR_LIGHT_SHADOWS ];
	#endif
	#if NUM_SPOT_LIGHT_SHADOWS > 0
		uniform sampler2D spotShadowMap[ NUM_SPOT_LIGHT_SHADOWS ];
		struct SpotLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
		};
		uniform SpotLightShadow spotLightShadows[ NUM_SPOT_LIGHT_SHADOWS ];
	#endif
	#if NUM_POINT_LIGHT_SHADOWS > 0
		uniform sampler2D pointShadowMap[ NUM_POINT_LIGHT_SHADOWS ];
		varying vec4 vPointShadowCoord[ NUM_POINT_LIGHT_SHADOWS ];
		struct PointLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
			float shadowCameraNear;
			float shadowCameraFar;
		};
		uniform PointLightShadow pointLightShadows[ NUM_POINT_LIGHT_SHADOWS ];
	#endif
	float texture2DCompare( sampler2D depths, vec2 uv, float compare ) {
		return step( compare, unpackRGBAToDepth( texture2D( depths, uv ) ) );
	}
	vec2 texture2DDistribution( sampler2D shadow, vec2 uv ) {
		return unpackRGBATo2Half( texture2D( shadow, uv ) );
	}
	float VSMShadow (sampler2D shadow, vec2 uv, float compare ){
		float occlusion = 1.0;
		vec2 distribution = texture2DDistribution( shadow, uv );
		float hard_shadow = step( compare , distribution.x );
		if (hard_shadow != 1.0 ) {
			float distance = compare - distribution.x ;
			float variance = max( 0.00000, distribution.y * distribution.y );
			float softness_probability = variance / (variance + distance * distance );			softness_probability = clamp( ( softness_probability - 0.3 ) / ( 0.95 - 0.3 ), 0.0, 1.0 );			occlusion = clamp( max( hard_shadow, softness_probability ), 0.0, 1.0 );
		}
		return occlusion;
	}
	float getShadow( sampler2D shadowMap, vec2 shadowMapSize, float shadowIntensity, float shadowBias, float shadowRadius, vec4 shadowCoord ) {
		float shadow = 1.0;
		shadowCoord.xyz /= shadowCoord.w;
		shadowCoord.z += shadowBias;
		bool inFrustum = shadowCoord.x >= 0.0 && shadowCoord.x <= 1.0 && shadowCoord.y >= 0.0 && shadowCoord.y <= 1.0;
		bool frustumTest = inFrustum && shadowCoord.z <= 1.0;
		if ( frustumTest ) {
		#if defined( SHADOWMAP_TYPE_PCF )
			vec2 texelSize = vec2( 1.0 ) / shadowMapSize;
			float dx0 = - texelSize.x * shadowRadius;
			float dy0 = - texelSize.y * shadowRadius;
			float dx1 = + texelSize.x * shadowRadius;
			float dy1 = + texelSize.y * shadowRadius;
			float dx2 = dx0 / 2.0;
			float dy2 = dy0 / 2.0;
			float dx3 = dx1 / 2.0;
			float dy3 = dy1 / 2.0;
			shadow = (
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx0, dy0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( 0.0, dy0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx1, dy0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx2, dy2 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( 0.0, dy2 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx3, dy2 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx0, 0.0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx2, 0.0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy, shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx3, 0.0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx1, 0.0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx2, dy3 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( 0.0, dy3 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx3, dy3 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx0, dy1 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( 0.0, dy1 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx1, dy1 ), shadowCoord.z )
			) * ( 1.0 / 17.0 );
		#elif defined( SHADOWMAP_TYPE_PCF_SOFT )
			vec2 texelSize = vec2( 1.0 ) / shadowMapSize;
			float dx = texelSize.x;
			float dy = texelSize.y;
			vec2 uv = shadowCoord.xy;
			vec2 f = fract( uv * shadowMapSize + 0.5 );
			uv -= f * texelSize;
			shadow = (
				texture2DCompare( shadowMap, uv, shadowCoord.z ) +
				texture2DCompare( shadowMap, uv + vec2( dx, 0.0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, uv + vec2( 0.0, dy ), shadowCoord.z ) +
				texture2DCompare( shadowMap, uv + texelSize, shadowCoord.z ) +
				mix( texture2DCompare( shadowMap, uv + vec2( -dx, 0.0 ), shadowCoord.z ),
					 texture2DCompare( shadowMap, uv + vec2( 2.0 * dx, 0.0 ), shadowCoord.z ),
					 f.x ) +
				mix( texture2DCompare( shadowMap, uv + vec2( -dx, dy ), shadowCoord.z ),
					 texture2DCompare( shadowMap, uv + vec2( 2.0 * dx, dy ), shadowCoord.z ),
					 f.x ) +
				mix( texture2DCompare( shadowMap, uv + vec2( 0.0, -dy ), shadowCoord.z ),
					 texture2DCompare( shadowMap, uv + vec2( 0.0, 2.0 * dy ), shadowCoord.z ),
					 f.y ) +
				mix( texture2DCompare( shadowMap, uv + vec2( dx, -dy ), shadowCoord.z ),
					 texture2DCompare( shadowMap, uv + vec2( dx, 2.0 * dy ), shadowCoord.z ),
					 f.y ) +
				mix( mix( texture2DCompare( shadowMap, uv + vec2( -dx, -dy ), shadowCoord.z ),
						  texture2DCompare( shadowMap, uv + vec2( 2.0 * dx, -dy ), shadowCoord.z ),
						  f.x ),
					 mix( texture2DCompare( shadowMap, uv + vec2( -dx, 2.0 * dy ), shadowCoord.z ),
						  texture2DCompare( shadowMap, uv + vec2( 2.0 * dx, 2.0 * dy ), shadowCoord.z ),
						  f.x ),
					 f.y )
			) * ( 1.0 / 9.0 );
		#elif defined( SHADOWMAP_TYPE_VSM )
			shadow = VSMShadow( shadowMap, shadowCoord.xy, shadowCoord.z );
		#else
			shadow = texture2DCompare( shadowMap, shadowCoord.xy, shadowCoord.z );
		#endif
		}
		return mix( 1.0, shadow, shadowIntensity );
	}
	vec2 cubeToUV( vec3 v, float texelSizeY ) {
		vec3 absV = abs( v );
		float scaleToCube = 1.0 / max( absV.x, max( absV.y, absV.z ) );
		absV *= scaleToCube;
		v *= scaleToCube * ( 1.0 - 2.0 * texelSizeY );
		vec2 planar = v.xy;
		float almostATexel = 1.5 * texelSizeY;
		float almostOne = 1.0 - almostATexel;
		if ( absV.z >= almostOne ) {
			if ( v.z > 0.0 )
				planar.x = 4.0 - v.x;
		} else if ( absV.x >= almostOne ) {
			float signX = sign( v.x );
			planar.x = v.z * signX + 2.0 * signX;
		} else if ( absV.y >= almostOne ) {
			float signY = sign( v.y );
			planar.x = v.x + 2.0 * signY + 2.0;
			planar.y = v.z * signY - 2.0;
		}
		return vec2( 0.125, 0.25 ) * planar + vec2( 0.375, 0.75 );
	}
	float getPointShadow( sampler2D shadowMap, vec2 shadowMapSize, float shadowIntensity, float shadowBias, float shadowRadius, vec4 shadowCoord, float shadowCameraNear, float shadowCameraFar ) {
		float shadow = 1.0;
		vec3 lightToPosition = shadowCoord.xyz;
		
		float lightToPositionLength = length( lightToPosition );
		if ( lightToPositionLength - shadowCameraFar <= 0.0 && lightToPositionLength - shadowCameraNear >= 0.0 ) {
			float dp = ( lightToPositionLength - shadowCameraNear ) / ( shadowCameraFar - shadowCameraNear );			dp += shadowBias;
			vec3 bd3D = normalize( lightToPosition );
			vec2 texelSize = vec2( 1.0 ) / ( shadowMapSize * vec2( 4.0, 2.0 ) );
			#if defined( SHADOWMAP_TYPE_PCF ) || defined( SHADOWMAP_TYPE_PCF_SOFT ) || defined( SHADOWMAP_TYPE_VSM )
				vec2 offset = vec2( - 1, 1 ) * shadowRadius * texelSize.y;
				shadow = (
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.xyy, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.yyy, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.xyx, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.yyx, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.xxy, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.yxy, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.xxx, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.yxx, texelSize.y ), dp )
				) * ( 1.0 / 9.0 );
			#else
				shadow = texture2DCompare( shadowMap, cubeToUV( bd3D, texelSize.y ), dp );
			#endif
		}
		return mix( 1.0, shadow, shadowIntensity );
	}
#endif`,mM=`#if NUM_SPOT_LIGHT_COORDS > 0
	uniform mat4 spotLightMatrix[ NUM_SPOT_LIGHT_COORDS ];
	varying vec4 vSpotLightCoord[ NUM_SPOT_LIGHT_COORDS ];
#endif
#ifdef USE_SHADOWMAP
	#if NUM_DIR_LIGHT_SHADOWS > 0
		uniform mat4 directionalShadowMatrix[ NUM_DIR_LIGHT_SHADOWS ];
		varying vec4 vDirectionalShadowCoord[ NUM_DIR_LIGHT_SHADOWS ];
		struct DirectionalLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
		};
		uniform DirectionalLightShadow directionalLightShadows[ NUM_DIR_LIGHT_SHADOWS ];
	#endif
	#if NUM_SPOT_LIGHT_SHADOWS > 0
		struct SpotLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
		};
		uniform SpotLightShadow spotLightShadows[ NUM_SPOT_LIGHT_SHADOWS ];
	#endif
	#if NUM_POINT_LIGHT_SHADOWS > 0
		uniform mat4 pointShadowMatrix[ NUM_POINT_LIGHT_SHADOWS ];
		varying vec4 vPointShadowCoord[ NUM_POINT_LIGHT_SHADOWS ];
		struct PointLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
			float shadowCameraNear;
			float shadowCameraFar;
		};
		uniform PointLightShadow pointLightShadows[ NUM_POINT_LIGHT_SHADOWS ];
	#endif
#endif`,gM=`#if ( defined( USE_SHADOWMAP ) && ( NUM_DIR_LIGHT_SHADOWS > 0 || NUM_POINT_LIGHT_SHADOWS > 0 ) ) || ( NUM_SPOT_LIGHT_COORDS > 0 )
	vec3 shadowWorldNormal = inverseTransformDirection( transformedNormal, viewMatrix );
	vec4 shadowWorldPosition;
#endif
#if defined( USE_SHADOWMAP )
	#if NUM_DIR_LIGHT_SHADOWS > 0
		#pragma unroll_loop_start
		for ( int i = 0; i < NUM_DIR_LIGHT_SHADOWS; i ++ ) {
			shadowWorldPosition = worldPosition + vec4( shadowWorldNormal * directionalLightShadows[ i ].shadowNormalBias, 0 );
			vDirectionalShadowCoord[ i ] = directionalShadowMatrix[ i ] * shadowWorldPosition;
		}
		#pragma unroll_loop_end
	#endif
	#if NUM_POINT_LIGHT_SHADOWS > 0
		#pragma unroll_loop_start
		for ( int i = 0; i < NUM_POINT_LIGHT_SHADOWS; i ++ ) {
			shadowWorldPosition = worldPosition + vec4( shadowWorldNormal * pointLightShadows[ i ].shadowNormalBias, 0 );
			vPointShadowCoord[ i ] = pointShadowMatrix[ i ] * shadowWorldPosition;
		}
		#pragma unroll_loop_end
	#endif
#endif
#if NUM_SPOT_LIGHT_COORDS > 0
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_SPOT_LIGHT_COORDS; i ++ ) {
		shadowWorldPosition = worldPosition;
		#if ( defined( USE_SHADOWMAP ) && UNROLLED_LOOP_INDEX < NUM_SPOT_LIGHT_SHADOWS )
			shadowWorldPosition.xyz += shadowWorldNormal * spotLightShadows[ i ].shadowNormalBias;
		#endif
		vSpotLightCoord[ i ] = spotLightMatrix[ i ] * shadowWorldPosition;
	}
	#pragma unroll_loop_end
#endif`,_M=`float getShadowMask() {
	float shadow = 1.0;
	#ifdef USE_SHADOWMAP
	#if NUM_DIR_LIGHT_SHADOWS > 0
	DirectionalLightShadow directionalLight;
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_DIR_LIGHT_SHADOWS; i ++ ) {
		directionalLight = directionalLightShadows[ i ];
		shadow *= receiveShadow ? getShadow( directionalShadowMap[ i ], directionalLight.shadowMapSize, directionalLight.shadowIntensity, directionalLight.shadowBias, directionalLight.shadowRadius, vDirectionalShadowCoord[ i ] ) : 1.0;
	}
	#pragma unroll_loop_end
	#endif
	#if NUM_SPOT_LIGHT_SHADOWS > 0
	SpotLightShadow spotLight;
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_SPOT_LIGHT_SHADOWS; i ++ ) {
		spotLight = spotLightShadows[ i ];
		shadow *= receiveShadow ? getShadow( spotShadowMap[ i ], spotLight.shadowMapSize, spotLight.shadowIntensity, spotLight.shadowBias, spotLight.shadowRadius, vSpotLightCoord[ i ] ) : 1.0;
	}
	#pragma unroll_loop_end
	#endif
	#if NUM_POINT_LIGHT_SHADOWS > 0
	PointLightShadow pointLight;
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_POINT_LIGHT_SHADOWS; i ++ ) {
		pointLight = pointLightShadows[ i ];
		shadow *= receiveShadow ? getPointShadow( pointShadowMap[ i ], pointLight.shadowMapSize, pointLight.shadowIntensity, pointLight.shadowBias, pointLight.shadowRadius, vPointShadowCoord[ i ], pointLight.shadowCameraNear, pointLight.shadowCameraFar ) : 1.0;
	}
	#pragma unroll_loop_end
	#endif
	#endif
	return shadow;
}`,vM=`#ifdef USE_SKINNING
	mat4 boneMatX = getBoneMatrix( skinIndex.x );
	mat4 boneMatY = getBoneMatrix( skinIndex.y );
	mat4 boneMatZ = getBoneMatrix( skinIndex.z );
	mat4 boneMatW = getBoneMatrix( skinIndex.w );
#endif`,xM=`#ifdef USE_SKINNING
	uniform mat4 bindMatrix;
	uniform mat4 bindMatrixInverse;
	uniform highp sampler2D boneTexture;
	mat4 getBoneMatrix( const in float i ) {
		int size = textureSize( boneTexture, 0 ).x;
		int j = int( i ) * 4;
		int x = j % size;
		int y = j / size;
		vec4 v1 = texelFetch( boneTexture, ivec2( x, y ), 0 );
		vec4 v2 = texelFetch( boneTexture, ivec2( x + 1, y ), 0 );
		vec4 v3 = texelFetch( boneTexture, ivec2( x + 2, y ), 0 );
		vec4 v4 = texelFetch( boneTexture, ivec2( x + 3, y ), 0 );
		return mat4( v1, v2, v3, v4 );
	}
#endif`,yM=`#ifdef USE_SKINNING
	vec4 skinVertex = bindMatrix * vec4( transformed, 1.0 );
	vec4 skinned = vec4( 0.0 );
	skinned += boneMatX * skinVertex * skinWeight.x;
	skinned += boneMatY * skinVertex * skinWeight.y;
	skinned += boneMatZ * skinVertex * skinWeight.z;
	skinned += boneMatW * skinVertex * skinWeight.w;
	transformed = ( bindMatrixInverse * skinned ).xyz;
#endif`,SM=`#ifdef USE_SKINNING
	mat4 skinMatrix = mat4( 0.0 );
	skinMatrix += skinWeight.x * boneMatX;
	skinMatrix += skinWeight.y * boneMatY;
	skinMatrix += skinWeight.z * boneMatZ;
	skinMatrix += skinWeight.w * boneMatW;
	skinMatrix = bindMatrixInverse * skinMatrix * bindMatrix;
	objectNormal = vec4( skinMatrix * vec4( objectNormal, 0.0 ) ).xyz;
	#ifdef USE_TANGENT
		objectTangent = vec4( skinMatrix * vec4( objectTangent, 0.0 ) ).xyz;
	#endif
#endif`,MM=`float specularStrength;
#ifdef USE_SPECULARMAP
	vec4 texelSpecular = texture2D( specularMap, vSpecularMapUv );
	specularStrength = texelSpecular.r;
#else
	specularStrength = 1.0;
#endif`,EM=`#ifdef USE_SPECULARMAP
	uniform sampler2D specularMap;
#endif`,TM=`#if defined( TONE_MAPPING )
	gl_FragColor.rgb = toneMapping( gl_FragColor.rgb );
#endif`,wM=`#ifndef saturate
#define saturate( a ) clamp( a, 0.0, 1.0 )
#endif
uniform float toneMappingExposure;
vec3 LinearToneMapping( vec3 color ) {
	return saturate( toneMappingExposure * color );
}
vec3 ReinhardToneMapping( vec3 color ) {
	color *= toneMappingExposure;
	return saturate( color / ( vec3( 1.0 ) + color ) );
}
vec3 CineonToneMapping( vec3 color ) {
	color *= toneMappingExposure;
	color = max( vec3( 0.0 ), color - 0.004 );
	return pow( ( color * ( 6.2 * color + 0.5 ) ) / ( color * ( 6.2 * color + 1.7 ) + 0.06 ), vec3( 2.2 ) );
}
vec3 RRTAndODTFit( vec3 v ) {
	vec3 a = v * ( v + 0.0245786 ) - 0.000090537;
	vec3 b = v * ( 0.983729 * v + 0.4329510 ) + 0.238081;
	return a / b;
}
vec3 ACESFilmicToneMapping( vec3 color ) {
	const mat3 ACESInputMat = mat3(
		vec3( 0.59719, 0.07600, 0.02840 ),		vec3( 0.35458, 0.90834, 0.13383 ),
		vec3( 0.04823, 0.01566, 0.83777 )
	);
	const mat3 ACESOutputMat = mat3(
		vec3(  1.60475, -0.10208, -0.00327 ),		vec3( -0.53108,  1.10813, -0.07276 ),
		vec3( -0.07367, -0.00605,  1.07602 )
	);
	color *= toneMappingExposure / 0.6;
	color = ACESInputMat * color;
	color = RRTAndODTFit( color );
	color = ACESOutputMat * color;
	return saturate( color );
}
const mat3 LINEAR_REC2020_TO_LINEAR_SRGB = mat3(
	vec3( 1.6605, - 0.1246, - 0.0182 ),
	vec3( - 0.5876, 1.1329, - 0.1006 ),
	vec3( - 0.0728, - 0.0083, 1.1187 )
);
const mat3 LINEAR_SRGB_TO_LINEAR_REC2020 = mat3(
	vec3( 0.6274, 0.0691, 0.0164 ),
	vec3( 0.3293, 0.9195, 0.0880 ),
	vec3( 0.0433, 0.0113, 0.8956 )
);
vec3 agxDefaultContrastApprox( vec3 x ) {
	vec3 x2 = x * x;
	vec3 x4 = x2 * x2;
	return + 15.5 * x4 * x2
		- 40.14 * x4 * x
		+ 31.96 * x4
		- 6.868 * x2 * x
		+ 0.4298 * x2
		+ 0.1191 * x
		- 0.00232;
}
vec3 AgXToneMapping( vec3 color ) {
	const mat3 AgXInsetMatrix = mat3(
		vec3( 0.856627153315983, 0.137318972929847, 0.11189821299995 ),
		vec3( 0.0951212405381588, 0.761241990602591, 0.0767994186031903 ),
		vec3( 0.0482516061458583, 0.101439036467562, 0.811302368396859 )
	);
	const mat3 AgXOutsetMatrix = mat3(
		vec3( 1.1271005818144368, - 0.1413297634984383, - 0.14132976349843826 ),
		vec3( - 0.11060664309660323, 1.157823702216272, - 0.11060664309660294 ),
		vec3( - 0.016493938717834573, - 0.016493938717834257, 1.2519364065950405 )
	);
	const float AgxMinEv = - 12.47393;	const float AgxMaxEv = 4.026069;
	color *= toneMappingExposure;
	color = LINEAR_SRGB_TO_LINEAR_REC2020 * color;
	color = AgXInsetMatrix * color;
	color = max( color, 1e-10 );	color = log2( color );
	color = ( color - AgxMinEv ) / ( AgxMaxEv - AgxMinEv );
	color = clamp( color, 0.0, 1.0 );
	color = agxDefaultContrastApprox( color );
	color = AgXOutsetMatrix * color;
	color = pow( max( vec3( 0.0 ), color ), vec3( 2.2 ) );
	color = LINEAR_REC2020_TO_LINEAR_SRGB * color;
	color = clamp( color, 0.0, 1.0 );
	return color;
}
vec3 NeutralToneMapping( vec3 color ) {
	const float StartCompression = 0.8 - 0.04;
	const float Desaturation = 0.15;
	color *= toneMappingExposure;
	float x = min( color.r, min( color.g, color.b ) );
	float offset = x < 0.08 ? x - 6.25 * x * x : 0.04;
	color -= offset;
	float peak = max( color.r, max( color.g, color.b ) );
	if ( peak < StartCompression ) return color;
	float d = 1. - StartCompression;
	float newPeak = 1. - d * d / ( peak + d - StartCompression );
	color *= newPeak / peak;
	float g = 1. - 1. / ( Desaturation * ( peak - newPeak ) + 1. );
	return mix( color, vec3( newPeak ), g );
}
vec3 CustomToneMapping( vec3 color ) { return color; }`,AM=`#ifdef USE_TRANSMISSION
	material.transmission = transmission;
	material.transmissionAlpha = 1.0;
	material.thickness = thickness;
	material.attenuationDistance = attenuationDistance;
	material.attenuationColor = attenuationColor;
	#ifdef USE_TRANSMISSIONMAP
		material.transmission *= texture2D( transmissionMap, vTransmissionMapUv ).r;
	#endif
	#ifdef USE_THICKNESSMAP
		material.thickness *= texture2D( thicknessMap, vThicknessMapUv ).g;
	#endif
	vec3 pos = vWorldPosition;
	vec3 v = normalize( cameraPosition - pos );
	vec3 n = inverseTransformDirection( normal, viewMatrix );
	vec4 transmitted = getIBLVolumeRefraction(
		n, v, material.roughness, material.diffuseColor, material.specularColor, material.specularF90,
		pos, modelMatrix, viewMatrix, projectionMatrix, material.dispersion, material.ior, material.thickness,
		material.attenuationColor, material.attenuationDistance );
	material.transmissionAlpha = mix( material.transmissionAlpha, transmitted.a, material.transmission );
	totalDiffuse = mix( totalDiffuse, transmitted.rgb, material.transmission );
#endif`,CM=`#ifdef USE_TRANSMISSION
	uniform float transmission;
	uniform float thickness;
	uniform float attenuationDistance;
	uniform vec3 attenuationColor;
	#ifdef USE_TRANSMISSIONMAP
		uniform sampler2D transmissionMap;
	#endif
	#ifdef USE_THICKNESSMAP
		uniform sampler2D thicknessMap;
	#endif
	uniform vec2 transmissionSamplerSize;
	uniform sampler2D transmissionSamplerMap;
	uniform mat4 modelMatrix;
	uniform mat4 projectionMatrix;
	varying vec3 vWorldPosition;
	float w0( float a ) {
		return ( 1.0 / 6.0 ) * ( a * ( a * ( - a + 3.0 ) - 3.0 ) + 1.0 );
	}
	float w1( float a ) {
		return ( 1.0 / 6.0 ) * ( a *  a * ( 3.0 * a - 6.0 ) + 4.0 );
	}
	float w2( float a ){
		return ( 1.0 / 6.0 ) * ( a * ( a * ( - 3.0 * a + 3.0 ) + 3.0 ) + 1.0 );
	}
	float w3( float a ) {
		return ( 1.0 / 6.0 ) * ( a * a * a );
	}
	float g0( float a ) {
		return w0( a ) + w1( a );
	}
	float g1( float a ) {
		return w2( a ) + w3( a );
	}
	float h0( float a ) {
		return - 1.0 + w1( a ) / ( w0( a ) + w1( a ) );
	}
	float h1( float a ) {
		return 1.0 + w3( a ) / ( w2( a ) + w3( a ) );
	}
	vec4 bicubic( sampler2D tex, vec2 uv, vec4 texelSize, float lod ) {
		uv = uv * texelSize.zw + 0.5;
		vec2 iuv = floor( uv );
		vec2 fuv = fract( uv );
		float g0x = g0( fuv.x );
		float g1x = g1( fuv.x );
		float h0x = h0( fuv.x );
		float h1x = h1( fuv.x );
		float h0y = h0( fuv.y );
		float h1y = h1( fuv.y );
		vec2 p0 = ( vec2( iuv.x + h0x, iuv.y + h0y ) - 0.5 ) * texelSize.xy;
		vec2 p1 = ( vec2( iuv.x + h1x, iuv.y + h0y ) - 0.5 ) * texelSize.xy;
		vec2 p2 = ( vec2( iuv.x + h0x, iuv.y + h1y ) - 0.5 ) * texelSize.xy;
		vec2 p3 = ( vec2( iuv.x + h1x, iuv.y + h1y ) - 0.5 ) * texelSize.xy;
		return g0( fuv.y ) * ( g0x * textureLod( tex, p0, lod ) + g1x * textureLod( tex, p1, lod ) ) +
			g1( fuv.y ) * ( g0x * textureLod( tex, p2, lod ) + g1x * textureLod( tex, p3, lod ) );
	}
	vec4 textureBicubic( sampler2D sampler, vec2 uv, float lod ) {
		vec2 fLodSize = vec2( textureSize( sampler, int( lod ) ) );
		vec2 cLodSize = vec2( textureSize( sampler, int( lod + 1.0 ) ) );
		vec2 fLodSizeInv = 1.0 / fLodSize;
		vec2 cLodSizeInv = 1.0 / cLodSize;
		vec4 fSample = bicubic( sampler, uv, vec4( fLodSizeInv, fLodSize ), floor( lod ) );
		vec4 cSample = bicubic( sampler, uv, vec4( cLodSizeInv, cLodSize ), ceil( lod ) );
		return mix( fSample, cSample, fract( lod ) );
	}
	vec3 getVolumeTransmissionRay( const in vec3 n, const in vec3 v, const in float thickness, const in float ior, const in mat4 modelMatrix ) {
		vec3 refractionVector = refract( - v, normalize( n ), 1.0 / ior );
		vec3 modelScale;
		modelScale.x = length( vec3( modelMatrix[ 0 ].xyz ) );
		modelScale.y = length( vec3( modelMatrix[ 1 ].xyz ) );
		modelScale.z = length( vec3( modelMatrix[ 2 ].xyz ) );
		return normalize( refractionVector ) * thickness * modelScale;
	}
	float applyIorToRoughness( const in float roughness, const in float ior ) {
		return roughness * clamp( ior * 2.0 - 2.0, 0.0, 1.0 );
	}
	vec4 getTransmissionSample( const in vec2 fragCoord, const in float roughness, const in float ior ) {
		float lod = log2( transmissionSamplerSize.x ) * applyIorToRoughness( roughness, ior );
		return textureBicubic( transmissionSamplerMap, fragCoord.xy, lod );
	}
	vec3 volumeAttenuation( const in float transmissionDistance, const in vec3 attenuationColor, const in float attenuationDistance ) {
		if ( isinf( attenuationDistance ) ) {
			return vec3( 1.0 );
		} else {
			vec3 attenuationCoefficient = -log( attenuationColor ) / attenuationDistance;
			vec3 transmittance = exp( - attenuationCoefficient * transmissionDistance );			return transmittance;
		}
	}
	vec4 getIBLVolumeRefraction( const in vec3 n, const in vec3 v, const in float roughness, const in vec3 diffuseColor,
		const in vec3 specularColor, const in float specularF90, const in vec3 position, const in mat4 modelMatrix,
		const in mat4 viewMatrix, const in mat4 projMatrix, const in float dispersion, const in float ior, const in float thickness,
		const in vec3 attenuationColor, const in float attenuationDistance ) {
		vec4 transmittedLight;
		vec3 transmittance;
		#ifdef USE_DISPERSION
			float halfSpread = ( ior - 1.0 ) * 0.025 * dispersion;
			vec3 iors = vec3( ior - halfSpread, ior, ior + halfSpread );
			for ( int i = 0; i < 3; i ++ ) {
				vec3 transmissionRay = getVolumeTransmissionRay( n, v, thickness, iors[ i ], modelMatrix );
				vec3 refractedRayExit = position + transmissionRay;
		
				vec4 ndcPos = projMatrix * viewMatrix * vec4( refractedRayExit, 1.0 );
				vec2 refractionCoords = ndcPos.xy / ndcPos.w;
				refractionCoords += 1.0;
				refractionCoords /= 2.0;
		
				vec4 transmissionSample = getTransmissionSample( refractionCoords, roughness, iors[ i ] );
				transmittedLight[ i ] = transmissionSample[ i ];
				transmittedLight.a += transmissionSample.a;
				transmittance[ i ] = diffuseColor[ i ] * volumeAttenuation( length( transmissionRay ), attenuationColor, attenuationDistance )[ i ];
			}
			transmittedLight.a /= 3.0;
		
		#else
		
			vec3 transmissionRay = getVolumeTransmissionRay( n, v, thickness, ior, modelMatrix );
			vec3 refractedRayExit = position + transmissionRay;
			vec4 ndcPos = projMatrix * viewMatrix * vec4( refractedRayExit, 1.0 );
			vec2 refractionCoords = ndcPos.xy / ndcPos.w;
			refractionCoords += 1.0;
			refractionCoords /= 2.0;
			transmittedLight = getTransmissionSample( refractionCoords, roughness, ior );
			transmittance = diffuseColor * volumeAttenuation( length( transmissionRay ), attenuationColor, attenuationDistance );
		
		#endif
		vec3 attenuatedColor = transmittance * transmittedLight.rgb;
		vec3 F = EnvironmentBRDF( n, v, specularColor, specularF90, roughness );
		float transmittanceFactor = ( transmittance.r + transmittance.g + transmittance.b ) / 3.0;
		return vec4( ( 1.0 - F ) * attenuatedColor, 1.0 - ( 1.0 - transmittedLight.a ) * transmittanceFactor );
	}
#endif`,RM=`#if defined( USE_UV ) || defined( USE_ANISOTROPY )
	varying vec2 vUv;
#endif
#ifdef USE_MAP
	varying vec2 vMapUv;
#endif
#ifdef USE_ALPHAMAP
	varying vec2 vAlphaMapUv;
#endif
#ifdef USE_LIGHTMAP
	varying vec2 vLightMapUv;
#endif
#ifdef USE_AOMAP
	varying vec2 vAoMapUv;
#endif
#ifdef USE_BUMPMAP
	varying vec2 vBumpMapUv;
#endif
#ifdef USE_NORMALMAP
	varying vec2 vNormalMapUv;
#endif
#ifdef USE_EMISSIVEMAP
	varying vec2 vEmissiveMapUv;
#endif
#ifdef USE_METALNESSMAP
	varying vec2 vMetalnessMapUv;
#endif
#ifdef USE_ROUGHNESSMAP
	varying vec2 vRoughnessMapUv;
#endif
#ifdef USE_ANISOTROPYMAP
	varying vec2 vAnisotropyMapUv;
#endif
#ifdef USE_CLEARCOATMAP
	varying vec2 vClearcoatMapUv;
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	varying vec2 vClearcoatNormalMapUv;
#endif
#ifdef USE_CLEARCOAT_ROUGHNESSMAP
	varying vec2 vClearcoatRoughnessMapUv;
#endif
#ifdef USE_IRIDESCENCEMAP
	varying vec2 vIridescenceMapUv;
#endif
#ifdef USE_IRIDESCENCE_THICKNESSMAP
	varying vec2 vIridescenceThicknessMapUv;
#endif
#ifdef USE_SHEEN_COLORMAP
	varying vec2 vSheenColorMapUv;
#endif
#ifdef USE_SHEEN_ROUGHNESSMAP
	varying vec2 vSheenRoughnessMapUv;
#endif
#ifdef USE_SPECULARMAP
	varying vec2 vSpecularMapUv;
#endif
#ifdef USE_SPECULAR_COLORMAP
	varying vec2 vSpecularColorMapUv;
#endif
#ifdef USE_SPECULAR_INTENSITYMAP
	varying vec2 vSpecularIntensityMapUv;
#endif
#ifdef USE_TRANSMISSIONMAP
	uniform mat3 transmissionMapTransform;
	varying vec2 vTransmissionMapUv;
#endif
#ifdef USE_THICKNESSMAP
	uniform mat3 thicknessMapTransform;
	varying vec2 vThicknessMapUv;
#endif`,bM=`#if defined( USE_UV ) || defined( USE_ANISOTROPY )
	varying vec2 vUv;
#endif
#ifdef USE_MAP
	uniform mat3 mapTransform;
	varying vec2 vMapUv;
#endif
#ifdef USE_ALPHAMAP
	uniform mat3 alphaMapTransform;
	varying vec2 vAlphaMapUv;
#endif
#ifdef USE_LIGHTMAP
	uniform mat3 lightMapTransform;
	varying vec2 vLightMapUv;
#endif
#ifdef USE_AOMAP
	uniform mat3 aoMapTransform;
	varying vec2 vAoMapUv;
#endif
#ifdef USE_BUMPMAP
	uniform mat3 bumpMapTransform;
	varying vec2 vBumpMapUv;
#endif
#ifdef USE_NORMALMAP
	uniform mat3 normalMapTransform;
	varying vec2 vNormalMapUv;
#endif
#ifdef USE_DISPLACEMENTMAP
	uniform mat3 displacementMapTransform;
	varying vec2 vDisplacementMapUv;
#endif
#ifdef USE_EMISSIVEMAP
	uniform mat3 emissiveMapTransform;
	varying vec2 vEmissiveMapUv;
#endif
#ifdef USE_METALNESSMAP
	uniform mat3 metalnessMapTransform;
	varying vec2 vMetalnessMapUv;
#endif
#ifdef USE_ROUGHNESSMAP
	uniform mat3 roughnessMapTransform;
	varying vec2 vRoughnessMapUv;
#endif
#ifdef USE_ANISOTROPYMAP
	uniform mat3 anisotropyMapTransform;
	varying vec2 vAnisotropyMapUv;
#endif
#ifdef USE_CLEARCOATMAP
	uniform mat3 clearcoatMapTransform;
	varying vec2 vClearcoatMapUv;
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	uniform mat3 clearcoatNormalMapTransform;
	varying vec2 vClearcoatNormalMapUv;
#endif
#ifdef USE_CLEARCOAT_ROUGHNESSMAP
	uniform mat3 clearcoatRoughnessMapTransform;
	varying vec2 vClearcoatRoughnessMapUv;
#endif
#ifdef USE_SHEEN_COLORMAP
	uniform mat3 sheenColorMapTransform;
	varying vec2 vSheenColorMapUv;
#endif
#ifdef USE_SHEEN_ROUGHNESSMAP
	uniform mat3 sheenRoughnessMapTransform;
	varying vec2 vSheenRoughnessMapUv;
#endif
#ifdef USE_IRIDESCENCEMAP
	uniform mat3 iridescenceMapTransform;
	varying vec2 vIridescenceMapUv;
#endif
#ifdef USE_IRIDESCENCE_THICKNESSMAP
	uniform mat3 iridescenceThicknessMapTransform;
	varying vec2 vIridescenceThicknessMapUv;
#endif
#ifdef USE_SPECULARMAP
	uniform mat3 specularMapTransform;
	varying vec2 vSpecularMapUv;
#endif
#ifdef USE_SPECULAR_COLORMAP
	uniform mat3 specularColorMapTransform;
	varying vec2 vSpecularColorMapUv;
#endif
#ifdef USE_SPECULAR_INTENSITYMAP
	uniform mat3 specularIntensityMapTransform;
	varying vec2 vSpecularIntensityMapUv;
#endif
#ifdef USE_TRANSMISSIONMAP
	uniform mat3 transmissionMapTransform;
	varying vec2 vTransmissionMapUv;
#endif
#ifdef USE_THICKNESSMAP
	uniform mat3 thicknessMapTransform;
	varying vec2 vThicknessMapUv;
#endif`,PM=`#if defined( USE_UV ) || defined( USE_ANISOTROPY )
	vUv = vec3( uv, 1 ).xy;
#endif
#ifdef USE_MAP
	vMapUv = ( mapTransform * vec3( MAP_UV, 1 ) ).xy;
#endif
#ifdef USE_ALPHAMAP
	vAlphaMapUv = ( alphaMapTransform * vec3( ALPHAMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_LIGHTMAP
	vLightMapUv = ( lightMapTransform * vec3( LIGHTMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_AOMAP
	vAoMapUv = ( aoMapTransform * vec3( AOMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_BUMPMAP
	vBumpMapUv = ( bumpMapTransform * vec3( BUMPMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_NORMALMAP
	vNormalMapUv = ( normalMapTransform * vec3( NORMALMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_DISPLACEMENTMAP
	vDisplacementMapUv = ( displacementMapTransform * vec3( DISPLACEMENTMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_EMISSIVEMAP
	vEmissiveMapUv = ( emissiveMapTransform * vec3( EMISSIVEMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_METALNESSMAP
	vMetalnessMapUv = ( metalnessMapTransform * vec3( METALNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_ROUGHNESSMAP
	vRoughnessMapUv = ( roughnessMapTransform * vec3( ROUGHNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_ANISOTROPYMAP
	vAnisotropyMapUv = ( anisotropyMapTransform * vec3( ANISOTROPYMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_CLEARCOATMAP
	vClearcoatMapUv = ( clearcoatMapTransform * vec3( CLEARCOATMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	vClearcoatNormalMapUv = ( clearcoatNormalMapTransform * vec3( CLEARCOAT_NORMALMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_CLEARCOAT_ROUGHNESSMAP
	vClearcoatRoughnessMapUv = ( clearcoatRoughnessMapTransform * vec3( CLEARCOAT_ROUGHNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_IRIDESCENCEMAP
	vIridescenceMapUv = ( iridescenceMapTransform * vec3( IRIDESCENCEMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_IRIDESCENCE_THICKNESSMAP
	vIridescenceThicknessMapUv = ( iridescenceThicknessMapTransform * vec3( IRIDESCENCE_THICKNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SHEEN_COLORMAP
	vSheenColorMapUv = ( sheenColorMapTransform * vec3( SHEEN_COLORMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SHEEN_ROUGHNESSMAP
	vSheenRoughnessMapUv = ( sheenRoughnessMapTransform * vec3( SHEEN_ROUGHNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SPECULARMAP
	vSpecularMapUv = ( specularMapTransform * vec3( SPECULARMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SPECULAR_COLORMAP
	vSpecularColorMapUv = ( specularColorMapTransform * vec3( SPECULAR_COLORMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SPECULAR_INTENSITYMAP
	vSpecularIntensityMapUv = ( specularIntensityMapTransform * vec3( SPECULAR_INTENSITYMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_TRANSMISSIONMAP
	vTransmissionMapUv = ( transmissionMapTransform * vec3( TRANSMISSIONMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_THICKNESSMAP
	vThicknessMapUv = ( thicknessMapTransform * vec3( THICKNESSMAP_UV, 1 ) ).xy;
#endif`,LM=`#if defined( USE_ENVMAP ) || defined( DISTANCE ) || defined ( USE_SHADOWMAP ) || defined ( USE_TRANSMISSION ) || NUM_SPOT_LIGHT_COORDS > 0
	vec4 worldPosition = vec4( transformed, 1.0 );
	#ifdef USE_BATCHING
		worldPosition = batchingMatrix * worldPosition;
	#endif
	#ifdef USE_INSTANCING
		worldPosition = instanceMatrix * worldPosition;
	#endif
	worldPosition = modelMatrix * worldPosition;
#endif`;const DM=`varying vec2 vUv;
uniform mat3 uvTransform;
void main() {
	vUv = ( uvTransform * vec3( uv, 1 ) ).xy;
	gl_Position = vec4( position.xy, 1.0, 1.0 );
}`,NM=`uniform sampler2D t2D;
uniform float backgroundIntensity;
varying vec2 vUv;
void main() {
	vec4 texColor = texture2D( t2D, vUv );
	#ifdef DECODE_VIDEO_TEXTURE
		texColor = vec4( mix( pow( texColor.rgb * 0.9478672986 + vec3( 0.0521327014 ), vec3( 2.4 ) ), texColor.rgb * 0.0773993808, vec3( lessThanEqual( texColor.rgb, vec3( 0.04045 ) ) ) ), texColor.w );
	#endif
	texColor.rgb *= backgroundIntensity;
	gl_FragColor = texColor;
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
}`,IM=`varying vec3 vWorldDirection;
#include <common>
void main() {
	vWorldDirection = transformDirection( position, modelMatrix );
	#include <begin_vertex>
	#include <project_vertex>
	gl_Position.z = gl_Position.w;
}`,UM=`#ifdef ENVMAP_TYPE_CUBE
	uniform samplerCube envMap;
#elif defined( ENVMAP_TYPE_CUBE_UV )
	uniform sampler2D envMap;
#endif
uniform float flipEnvMap;
uniform float backgroundBlurriness;
uniform float backgroundIntensity;
uniform mat3 backgroundRotation;
varying vec3 vWorldDirection;
#include <cube_uv_reflection_fragment>
void main() {
	#ifdef ENVMAP_TYPE_CUBE
		vec4 texColor = textureCube( envMap, backgroundRotation * vec3( flipEnvMap * vWorldDirection.x, vWorldDirection.yz ) );
	#elif defined( ENVMAP_TYPE_CUBE_UV )
		vec4 texColor = textureCubeUV( envMap, backgroundRotation * vWorldDirection, backgroundBlurriness );
	#else
		vec4 texColor = vec4( 0.0, 0.0, 0.0, 1.0 );
	#endif
	texColor.rgb *= backgroundIntensity;
	gl_FragColor = texColor;
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
}`,FM=`varying vec3 vWorldDirection;
#include <common>
void main() {
	vWorldDirection = transformDirection( position, modelMatrix );
	#include <begin_vertex>
	#include <project_vertex>
	gl_Position.z = gl_Position.w;
}`,OM=`uniform samplerCube tCube;
uniform float tFlip;
uniform float opacity;
varying vec3 vWorldDirection;
void main() {
	vec4 texColor = textureCube( tCube, vec3( tFlip * vWorldDirection.x, vWorldDirection.yz ) );
	gl_FragColor = texColor;
	gl_FragColor.a *= opacity;
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
}`,kM=`#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
varying vec2 vHighPrecisionZW;
void main() {
	#include <uv_vertex>
	#include <batching_vertex>
	#include <skinbase_vertex>
	#include <morphinstance_vertex>
	#ifdef USE_DISPLACEMENTMAP
		#include <beginnormal_vertex>
		#include <morphnormal_vertex>
		#include <skinnormal_vertex>
	#endif
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vHighPrecisionZW = gl_Position.zw;
}`,BM=`#if DEPTH_PACKING == 3200
	uniform float opacity;
#endif
#include <common>
#include <packing>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
varying vec2 vHighPrecisionZW;
void main() {
	vec4 diffuseColor = vec4( 1.0 );
	#include <clipping_planes_fragment>
	#if DEPTH_PACKING == 3200
		diffuseColor.a = opacity;
	#endif
	#include <map_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <logdepthbuf_fragment>
	float fragCoordZ = 0.5 * vHighPrecisionZW[0] / vHighPrecisionZW[1] + 0.5;
	#if DEPTH_PACKING == 3200
		gl_FragColor = vec4( vec3( 1.0 - fragCoordZ ), opacity );
	#elif DEPTH_PACKING == 3201
		gl_FragColor = packDepthToRGBA( fragCoordZ );
	#elif DEPTH_PACKING == 3202
		gl_FragColor = vec4( packDepthToRGB( fragCoordZ ), 1.0 );
	#elif DEPTH_PACKING == 3203
		gl_FragColor = vec4( packDepthToRG( fragCoordZ ), 0.0, 1.0 );
	#endif
}`,zM=`#define DISTANCE
varying vec3 vWorldPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <batching_vertex>
	#include <skinbase_vertex>
	#include <morphinstance_vertex>
	#ifdef USE_DISPLACEMENTMAP
		#include <beginnormal_vertex>
		#include <morphnormal_vertex>
		#include <skinnormal_vertex>
	#endif
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <worldpos_vertex>
	#include <clipping_planes_vertex>
	vWorldPosition = worldPosition.xyz;
}`,HM=`#define DISTANCE
uniform vec3 referencePosition;
uniform float nearDistance;
uniform float farDistance;
varying vec3 vWorldPosition;
#include <common>
#include <packing>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <clipping_planes_pars_fragment>
void main () {
	vec4 diffuseColor = vec4( 1.0 );
	#include <clipping_planes_fragment>
	#include <map_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	float dist = length( vWorldPosition - referencePosition );
	dist = ( dist - nearDistance ) / ( farDistance - nearDistance );
	dist = saturate( dist );
	gl_FragColor = packDepthToRGBA( dist );
}`,VM=`varying vec3 vWorldDirection;
#include <common>
void main() {
	vWorldDirection = transformDirection( position, modelMatrix );
	#include <begin_vertex>
	#include <project_vertex>
}`,GM=`uniform sampler2D tEquirect;
varying vec3 vWorldDirection;
#include <common>
void main() {
	vec3 direction = normalize( vWorldDirection );
	vec2 sampleUV = equirectUv( direction );
	gl_FragColor = texture2D( tEquirect, sampleUV );
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
}`,WM=`uniform float scale;
attribute float lineDistance;
varying float vLineDistance;
#include <common>
#include <uv_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <morphtarget_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	vLineDistance = scale * lineDistance;
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <fog_vertex>
}`,jM=`uniform vec3 diffuse;
uniform float opacity;
uniform float dashSize;
uniform float totalSize;
varying float vLineDistance;
#include <common>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <fog_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	if ( mod( vLineDistance, totalSize ) > dashSize ) {
		discard;
	}
	vec3 outgoingLight = vec3( 0.0 );
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	outgoingLight = diffuseColor.rgb;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
}`,XM=`#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <envmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#if defined ( USE_ENVMAP ) || defined ( USE_SKINNING )
		#include <beginnormal_vertex>
		#include <morphnormal_vertex>
		#include <skinbase_vertex>
		#include <skinnormal_vertex>
		#include <defaultnormal_vertex>
	#endif
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <worldpos_vertex>
	#include <envmap_vertex>
	#include <fog_vertex>
}`,YM=`uniform vec3 diffuse;
uniform float opacity;
#ifndef FLAT_SHADED
	varying vec3 vNormal;
#endif
#include <common>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <envmap_common_pars_fragment>
#include <envmap_pars_fragment>
#include <fog_pars_fragment>
#include <specularmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <specularmap_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	#ifdef USE_LIGHTMAP
		vec4 lightMapTexel = texture2D( lightMap, vLightMapUv );
		reflectedLight.indirectDiffuse += lightMapTexel.rgb * lightMapIntensity * RECIPROCAL_PI;
	#else
		reflectedLight.indirectDiffuse += vec3( 1.0 );
	#endif
	#include <aomap_fragment>
	reflectedLight.indirectDiffuse *= diffuseColor.rgb;
	vec3 outgoingLight = reflectedLight.indirectDiffuse;
	#include <envmap_fragment>
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,qM=`#define LAMBERT
varying vec3 vViewPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <envmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <shadowmap_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vViewPosition = - mvPosition.xyz;
	#include <worldpos_vertex>
	#include <envmap_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
}`,$M=`#define LAMBERT
uniform vec3 diffuse;
uniform vec3 emissive;
uniform float opacity;
#include <common>
#include <packing>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <emissivemap_pars_fragment>
#include <envmap_common_pars_fragment>
#include <envmap_pars_fragment>
#include <fog_pars_fragment>
#include <bsdfs>
#include <lights_pars_begin>
#include <normal_pars_fragment>
#include <lights_lambert_pars_fragment>
#include <shadowmap_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <specularmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	vec3 totalEmissiveRadiance = emissive;
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <specularmap_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	#include <emissivemap_fragment>
	#include <lights_lambert_fragment>
	#include <lights_fragment_begin>
	#include <lights_fragment_maps>
	#include <lights_fragment_end>
	#include <aomap_fragment>
	vec3 outgoingLight = reflectedLight.directDiffuse + reflectedLight.indirectDiffuse + totalEmissiveRadiance;
	#include <envmap_fragment>
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,KM=`#define MATCAP
varying vec3 vViewPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <color_pars_vertex>
#include <displacementmap_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <fog_vertex>
	vViewPosition = - mvPosition.xyz;
}`,ZM=`#define MATCAP
uniform vec3 diffuse;
uniform float opacity;
uniform sampler2D matcap;
varying vec3 vViewPosition;
#include <common>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <fog_pars_fragment>
#include <normal_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	vec3 viewDir = normalize( vViewPosition );
	vec3 x = normalize( vec3( viewDir.z, 0.0, - viewDir.x ) );
	vec3 y = cross( viewDir, x );
	vec2 uv = vec2( dot( x, normal ), dot( y, normal ) ) * 0.495 + 0.5;
	#ifdef USE_MATCAP
		vec4 matcapColor = texture2D( matcap, uv );
	#else
		vec4 matcapColor = vec4( vec3( mix( 0.2, 0.8, uv.y ) ), 1.0 );
	#endif
	vec3 outgoingLight = diffuseColor.rgb * matcapColor.rgb;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,QM=`#define NORMAL
#if defined( FLAT_SHADED ) || defined( USE_BUMPMAP ) || defined( USE_NORMALMAP_TANGENTSPACE )
	varying vec3 vViewPosition;
#endif
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphinstance_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
#if defined( FLAT_SHADED ) || defined( USE_BUMPMAP ) || defined( USE_NORMALMAP_TANGENTSPACE )
	vViewPosition = - mvPosition.xyz;
#endif
}`,JM=`#define NORMAL
uniform float opacity;
#if defined( FLAT_SHADED ) || defined( USE_BUMPMAP ) || defined( USE_NORMALMAP_TANGENTSPACE )
	varying vec3 vViewPosition;
#endif
#include <packing>
#include <uv_pars_fragment>
#include <normal_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( 0.0, 0.0, 0.0, opacity );
	#include <clipping_planes_fragment>
	#include <logdepthbuf_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	gl_FragColor = vec4( packNormalToRGB( normal ), diffuseColor.a );
	#ifdef OPAQUE
		gl_FragColor.a = 1.0;
	#endif
}`,eE=`#define PHONG
varying vec3 vViewPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <envmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <shadowmap_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphinstance_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vViewPosition = - mvPosition.xyz;
	#include <worldpos_vertex>
	#include <envmap_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
}`,tE=`#define PHONG
uniform vec3 diffuse;
uniform vec3 emissive;
uniform vec3 specular;
uniform float shininess;
uniform float opacity;
#include <common>
#include <packing>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <emissivemap_pars_fragment>
#include <envmap_common_pars_fragment>
#include <envmap_pars_fragment>
#include <fog_pars_fragment>
#include <bsdfs>
#include <lights_pars_begin>
#include <normal_pars_fragment>
#include <lights_phong_pars_fragment>
#include <shadowmap_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <specularmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	vec3 totalEmissiveRadiance = emissive;
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <specularmap_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	#include <emissivemap_fragment>
	#include <lights_phong_fragment>
	#include <lights_fragment_begin>
	#include <lights_fragment_maps>
	#include <lights_fragment_end>
	#include <aomap_fragment>
	vec3 outgoingLight = reflectedLight.directDiffuse + reflectedLight.indirectDiffuse + reflectedLight.directSpecular + reflectedLight.indirectSpecular + totalEmissiveRadiance;
	#include <envmap_fragment>
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,nE=`#define STANDARD
varying vec3 vViewPosition;
#ifdef USE_TRANSMISSION
	varying vec3 vWorldPosition;
#endif
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <shadowmap_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vViewPosition = - mvPosition.xyz;
	#include <worldpos_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
#ifdef USE_TRANSMISSION
	vWorldPosition = worldPosition.xyz;
#endif
}`,iE=`#define STANDARD
#ifdef PHYSICAL
	#define IOR
	#define USE_SPECULAR
#endif
uniform vec3 diffuse;
uniform vec3 emissive;
uniform float roughness;
uniform float metalness;
uniform float opacity;
#ifdef IOR
	uniform float ior;
#endif
#ifdef USE_SPECULAR
	uniform float specularIntensity;
	uniform vec3 specularColor;
	#ifdef USE_SPECULAR_COLORMAP
		uniform sampler2D specularColorMap;
	#endif
	#ifdef USE_SPECULAR_INTENSITYMAP
		uniform sampler2D specularIntensityMap;
	#endif
#endif
#ifdef USE_CLEARCOAT
	uniform float clearcoat;
	uniform float clearcoatRoughness;
#endif
#ifdef USE_DISPERSION
	uniform float dispersion;
#endif
#ifdef USE_IRIDESCENCE
	uniform float iridescence;
	uniform float iridescenceIOR;
	uniform float iridescenceThicknessMinimum;
	uniform float iridescenceThicknessMaximum;
#endif
#ifdef USE_SHEEN
	uniform vec3 sheenColor;
	uniform float sheenRoughness;
	#ifdef USE_SHEEN_COLORMAP
		uniform sampler2D sheenColorMap;
	#endif
	#ifdef USE_SHEEN_ROUGHNESSMAP
		uniform sampler2D sheenRoughnessMap;
	#endif
#endif
#ifdef USE_ANISOTROPY
	uniform vec2 anisotropyVector;
	#ifdef USE_ANISOTROPYMAP
		uniform sampler2D anisotropyMap;
	#endif
#endif
varying vec3 vViewPosition;
#include <common>
#include <packing>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <emissivemap_pars_fragment>
#include <iridescence_fragment>
#include <cube_uv_reflection_fragment>
#include <envmap_common_pars_fragment>
#include <envmap_physical_pars_fragment>
#include <fog_pars_fragment>
#include <lights_pars_begin>
#include <normal_pars_fragment>
#include <lights_physical_pars_fragment>
#include <transmission_pars_fragment>
#include <shadowmap_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <clearcoat_pars_fragment>
#include <iridescence_pars_fragment>
#include <roughnessmap_pars_fragment>
#include <metalnessmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	vec3 totalEmissiveRadiance = emissive;
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <roughnessmap_fragment>
	#include <metalnessmap_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	#include <clearcoat_normal_fragment_begin>
	#include <clearcoat_normal_fragment_maps>
	#include <emissivemap_fragment>
	#include <lights_physical_fragment>
	#include <lights_fragment_begin>
	#include <lights_fragment_maps>
	#include <lights_fragment_end>
	#include <aomap_fragment>
	vec3 totalDiffuse = reflectedLight.directDiffuse + reflectedLight.indirectDiffuse;
	vec3 totalSpecular = reflectedLight.directSpecular + reflectedLight.indirectSpecular;
	#include <transmission_fragment>
	vec3 outgoingLight = totalDiffuse + totalSpecular + totalEmissiveRadiance;
	#ifdef USE_SHEEN
		float sheenEnergyComp = 1.0 - 0.157 * max3( material.sheenColor );
		outgoingLight = outgoingLight * sheenEnergyComp + sheenSpecularDirect + sheenSpecularIndirect;
	#endif
	#ifdef USE_CLEARCOAT
		float dotNVcc = saturate( dot( geometryClearcoatNormal, geometryViewDir ) );
		vec3 Fcc = F_Schlick( material.clearcoatF0, material.clearcoatF90, dotNVcc );
		outgoingLight = outgoingLight * ( 1.0 - material.clearcoat * Fcc ) + ( clearcoatSpecularDirect + clearcoatSpecularIndirect ) * material.clearcoat;
	#endif
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,rE=`#define TOON
varying vec3 vViewPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <shadowmap_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vViewPosition = - mvPosition.xyz;
	#include <worldpos_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
}`,sE=`#define TOON
uniform vec3 diffuse;
uniform vec3 emissive;
uniform float opacity;
#include <common>
#include <packing>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <emissivemap_pars_fragment>
#include <gradientmap_pars_fragment>
#include <fog_pars_fragment>
#include <bsdfs>
#include <lights_pars_begin>
#include <normal_pars_fragment>
#include <lights_toon_pars_fragment>
#include <shadowmap_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	vec3 totalEmissiveRadiance = emissive;
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	#include <emissivemap_fragment>
	#include <lights_toon_fragment>
	#include <lights_fragment_begin>
	#include <lights_fragment_maps>
	#include <lights_fragment_end>
	#include <aomap_fragment>
	vec3 outgoingLight = reflectedLight.directDiffuse + reflectedLight.indirectDiffuse + totalEmissiveRadiance;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,oE=`uniform float size;
uniform float scale;
#include <common>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <morphtarget_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
#ifdef USE_POINTS_UV
	varying vec2 vUv;
	uniform mat3 uvTransform;
#endif
void main() {
	#ifdef USE_POINTS_UV
		vUv = ( uvTransform * vec3( uv, 1 ) ).xy;
	#endif
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <project_vertex>
	gl_PointSize = size;
	#ifdef USE_SIZEATTENUATION
		bool isPerspective = isPerspectiveMatrix( projectionMatrix );
		if ( isPerspective ) gl_PointSize *= ( scale / - mvPosition.z );
	#endif
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <worldpos_vertex>
	#include <fog_vertex>
}`,aE=`uniform vec3 diffuse;
uniform float opacity;
#include <common>
#include <color_pars_fragment>
#include <map_particle_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <fog_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	vec3 outgoingLight = vec3( 0.0 );
	#include <logdepthbuf_fragment>
	#include <map_particle_fragment>
	#include <color_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	outgoingLight = diffuseColor.rgb;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
}`,lE=`#include <common>
#include <batching_pars_vertex>
#include <fog_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <shadowmap_pars_vertex>
void main() {
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphinstance_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <worldpos_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
}`,cE=`uniform vec3 color;
uniform float opacity;
#include <common>
#include <packing>
#include <fog_pars_fragment>
#include <bsdfs>
#include <lights_pars_begin>
#include <logdepthbuf_pars_fragment>
#include <shadowmap_pars_fragment>
#include <shadowmask_pars_fragment>
void main() {
	#include <logdepthbuf_fragment>
	gl_FragColor = vec4( color, opacity * ( 1.0 - getShadowMask() ) );
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
}`,uE=`uniform float rotation;
uniform vec2 center;
#include <common>
#include <uv_pars_vertex>
#include <fog_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	vec4 mvPosition = modelViewMatrix[ 3 ];
	vec2 scale = vec2( length( modelMatrix[ 0 ].xyz ), length( modelMatrix[ 1 ].xyz ) );
	#ifndef USE_SIZEATTENUATION
		bool isPerspective = isPerspectiveMatrix( projectionMatrix );
		if ( isPerspective ) scale *= - mvPosition.z;
	#endif
	vec2 alignedPosition = ( position.xy - ( center - vec2( 0.5 ) ) ) * scale;
	vec2 rotatedPosition;
	rotatedPosition.x = cos( rotation ) * alignedPosition.x - sin( rotation ) * alignedPosition.y;
	rotatedPosition.y = sin( rotation ) * alignedPosition.x + cos( rotation ) * alignedPosition.y;
	mvPosition.xy += rotatedPosition;
	gl_Position = projectionMatrix * mvPosition;
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <fog_vertex>
}`,fE=`uniform vec3 diffuse;
uniform float opacity;
#include <common>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <fog_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	vec3 outgoingLight = vec3( 0.0 );
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	outgoingLight = diffuseColor.rgb;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
}`,st={alphahash_fragment:Ny,alphahash_pars_fragment:Iy,alphamap_fragment:Uy,alphamap_pars_fragment:Fy,alphatest_fragment:Oy,alphatest_pars_fragment:ky,aomap_fragment:By,aomap_pars_fragment:zy,batching_pars_vertex:Hy,batching_vertex:Vy,begin_vertex:Gy,beginnormal_vertex:Wy,bsdfs:jy,iridescence_fragment:Xy,bumpmap_pars_fragment:Yy,clipping_planes_fragment:qy,clipping_planes_pars_fragment:$y,clipping_planes_pars_vertex:Ky,clipping_planes_vertex:Zy,color_fragment:Qy,color_pars_fragment:Jy,color_pars_vertex:eS,color_vertex:tS,common:nS,cube_uv_reflection_fragment:iS,defaultnormal_vertex:rS,displacementmap_pars_vertex:sS,displacementmap_vertex:oS,emissivemap_fragment:aS,emissivemap_pars_fragment:lS,colorspace_fragment:cS,colorspace_pars_fragment:uS,envmap_fragment:fS,envmap_common_pars_fragment:dS,envmap_pars_fragment:hS,envmap_pars_vertex:pS,envmap_physical_pars_fragment:wS,envmap_vertex:mS,fog_vertex:gS,fog_pars_vertex:_S,fog_fragment:vS,fog_pars_fragment:xS,gradientmap_pars_fragment:yS,lightmap_pars_fragment:SS,lights_lambert_fragment:MS,lights_lambert_pars_fragment:ES,lights_pars_begin:TS,lights_toon_fragment:AS,lights_toon_pars_fragment:CS,lights_phong_fragment:RS,lights_phong_pars_fragment:bS,lights_physical_fragment:PS,lights_physical_pars_fragment:LS,lights_fragment_begin:DS,lights_fragment_maps:NS,lights_fragment_end:IS,logdepthbuf_fragment:US,logdepthbuf_pars_fragment:FS,logdepthbuf_pars_vertex:OS,logdepthbuf_vertex:kS,map_fragment:BS,map_pars_fragment:zS,map_particle_fragment:HS,map_particle_pars_fragment:VS,metalnessmap_fragment:GS,metalnessmap_pars_fragment:WS,morphinstance_vertex:jS,morphcolor_vertex:XS,morphnormal_vertex:YS,morphtarget_pars_vertex:qS,morphtarget_vertex:$S,normal_fragment_begin:KS,normal_fragment_maps:ZS,normal_pars_fragment:QS,normal_pars_vertex:JS,normal_vertex:eM,normalmap_pars_fragment:tM,clearcoat_normal_fragment_begin:nM,clearcoat_normal_fragment_maps:iM,clearcoat_pars_fragment:rM,iridescence_pars_fragment:sM,opaque_fragment:oM,packing:aM,premultiplied_alpha_fragment:lM,project_vertex:cM,dithering_fragment:uM,dithering_pars_fragment:fM,roughnessmap_fragment:dM,roughnessmap_pars_fragment:hM,shadowmap_pars_fragment:pM,shadowmap_pars_vertex:mM,shadowmap_vertex:gM,shadowmask_pars_fragment:_M,skinbase_vertex:vM,skinning_pars_vertex:xM,skinning_vertex:yM,skinnormal_vertex:SM,specularmap_fragment:MM,specularmap_pars_fragment:EM,tonemapping_fragment:TM,tonemapping_pars_fragment:wM,transmission_fragment:AM,transmission_pars_fragment:CM,uv_pars_fragment:RM,uv_pars_vertex:bM,uv_vertex:PM,worldpos_vertex:LM,background_vert:DM,background_frag:NM,backgroundCube_vert:IM,backgroundCube_frag:UM,cube_vert:FM,cube_frag:OM,depth_vert:kM,depth_frag:BM,distanceRGBA_vert:zM,distanceRGBA_frag:HM,equirect_vert:VM,equirect_frag:GM,linedashed_vert:WM,linedashed_frag:jM,meshbasic_vert:XM,meshbasic_frag:YM,meshlambert_vert:qM,meshlambert_frag:$M,meshmatcap_vert:KM,meshmatcap_frag:ZM,meshnormal_vert:QM,meshnormal_frag:JM,meshphong_vert:eE,meshphong_frag:tE,meshphysical_vert:nE,meshphysical_frag:iE,meshtoon_vert:rE,meshtoon_frag:sE,points_vert:oE,points_frag:aE,shadow_vert:lE,shadow_frag:cE,sprite_vert:uE,sprite_frag:fE},be={common:{diffuse:{value:new vt(16777215)},opacity:{value:1},map:{value:null},mapTransform:{value:new ot},alphaMap:{value:null},alphaMapTransform:{value:new ot},alphaTest:{value:0}},specularmap:{specularMap:{value:null},specularMapTransform:{value:new ot}},envmap:{envMap:{value:null},envMapRotation:{value:new ot},flipEnvMap:{value:-1},reflectivity:{value:1},ior:{value:1.5},refractionRatio:{value:.98}},aomap:{aoMap:{value:null},aoMapIntensity:{value:1},aoMapTransform:{value:new ot}},lightmap:{lightMap:{value:null},lightMapIntensity:{value:1},lightMapTransform:{value:new ot}},bumpmap:{bumpMap:{value:null},bumpMapTransform:{value:new ot},bumpScale:{value:1}},normalmap:{normalMap:{value:null},normalMapTransform:{value:new ot},normalScale:{value:new rt(1,1)}},displacementmap:{displacementMap:{value:null},displacementMapTransform:{value:new ot},displacementScale:{value:1},displacementBias:{value:0}},emissivemap:{emissiveMap:{value:null},emissiveMapTransform:{value:new ot}},metalnessmap:{metalnessMap:{value:null},metalnessMapTransform:{value:new ot}},roughnessmap:{roughnessMap:{value:null},roughnessMapTransform:{value:new ot}},gradientmap:{gradientMap:{value:null}},fog:{fogDensity:{value:25e-5},fogNear:{value:1},fogFar:{value:2e3},fogColor:{value:new vt(16777215)}},lights:{ambientLightColor:{value:[]},lightProbe:{value:[]},directionalLights:{value:[],properties:{direction:{},color:{}}},directionalLightShadows:{value:[],properties:{shadowIntensity:1,shadowBias:{},shadowNormalBias:{},shadowRadius:{},shadowMapSize:{}}},directionalShadowMap:{value:[]},directionalShadowMatrix:{value:[]},spotLights:{value:[],properties:{color:{},position:{},direction:{},distance:{},coneCos:{},penumbraCos:{},decay:{}}},spotLightShadows:{value:[],properties:{shadowIntensity:1,shadowBias:{},shadowNormalBias:{},shadowRadius:{},shadowMapSize:{}}},spotLightMap:{value:[]},spotShadowMap:{value:[]},spotLightMatrix:{value:[]},pointLights:{value:[],properties:{color:{},position:{},decay:{},distance:{}}},pointLightShadows:{value:[],properties:{shadowIntensity:1,shadowBias:{},shadowNormalBias:{},shadowRadius:{},shadowMapSize:{},shadowCameraNear:{},shadowCameraFar:{}}},pointShadowMap:{value:[]},pointShadowMatrix:{value:[]},hemisphereLights:{value:[],properties:{direction:{},skyColor:{},groundColor:{}}},rectAreaLights:{value:[],properties:{color:{},position:{},width:{},height:{}}},ltc_1:{value:null},ltc_2:{value:null}},points:{diffuse:{value:new vt(16777215)},opacity:{value:1},size:{value:1},scale:{value:1},map:{value:null},alphaMap:{value:null},alphaMapTransform:{value:new ot},alphaTest:{value:0},uvTransform:{value:new ot}},sprite:{diffuse:{value:new vt(16777215)},opacity:{value:1},center:{value:new rt(.5,.5)},rotation:{value:0},map:{value:null},mapTransform:{value:new ot},alphaMap:{value:null},alphaMapTransform:{value:new ot},alphaTest:{value:0}}},vi={basic:{uniforms:Sn([be.common,be.specularmap,be.envmap,be.aomap,be.lightmap,be.fog]),vertexShader:st.meshbasic_vert,fragmentShader:st.meshbasic_frag},lambert:{uniforms:Sn([be.common,be.specularmap,be.envmap,be.aomap,be.lightmap,be.emissivemap,be.bumpmap,be.normalmap,be.displacementmap,be.fog,be.lights,{emissive:{value:new vt(0)}}]),vertexShader:st.meshlambert_vert,fragmentShader:st.meshlambert_frag},phong:{uniforms:Sn([be.common,be.specularmap,be.envmap,be.aomap,be.lightmap,be.emissivemap,be.bumpmap,be.normalmap,be.displacementmap,be.fog,be.lights,{emissive:{value:new vt(0)},specular:{value:new vt(1118481)},shininess:{value:30}}]),vertexShader:st.meshphong_vert,fragmentShader:st.meshphong_frag},standard:{uniforms:Sn([be.common,be.envmap,be.aomap,be.lightmap,be.emissivemap,be.bumpmap,be.normalmap,be.displacementmap,be.roughnessmap,be.metalnessmap,be.fog,be.lights,{emissive:{value:new vt(0)},roughness:{value:1},metalness:{value:0},envMapIntensity:{value:1}}]),vertexShader:st.meshphysical_vert,fragmentShader:st.meshphysical_frag},toon:{uniforms:Sn([be.common,be.aomap,be.lightmap,be.emissivemap,be.bumpmap,be.normalmap,be.displacementmap,be.gradientmap,be.fog,be.lights,{emissive:{value:new vt(0)}}]),vertexShader:st.meshtoon_vert,fragmentShader:st.meshtoon_frag},matcap:{uniforms:Sn([be.common,be.bumpmap,be.normalmap,be.displacementmap,be.fog,{matcap:{value:null}}]),vertexShader:st.meshmatcap_vert,fragmentShader:st.meshmatcap_frag},points:{uniforms:Sn([be.points,be.fog]),vertexShader:st.points_vert,fragmentShader:st.points_frag},dashed:{uniforms:Sn([be.common,be.fog,{scale:{value:1},dashSize:{value:1},totalSize:{value:2}}]),vertexShader:st.linedashed_vert,fragmentShader:st.linedashed_frag},depth:{uniforms:Sn([be.common,be.displacementmap]),vertexShader:st.depth_vert,fragmentShader:st.depth_frag},normal:{uniforms:Sn([be.common,be.bumpmap,be.normalmap,be.displacementmap,{opacity:{value:1}}]),vertexShader:st.meshnormal_vert,fragmentShader:st.meshnormal_frag},sprite:{uniforms:Sn([be.sprite,be.fog]),vertexShader:st.sprite_vert,fragmentShader:st.sprite_frag},background:{uniforms:{uvTransform:{value:new ot},t2D:{value:null},backgroundIntensity:{value:1}},vertexShader:st.background_vert,fragmentShader:st.background_frag},backgroundCube:{uniforms:{envMap:{value:null},flipEnvMap:{value:-1},backgroundBlurriness:{value:0},backgroundIntensity:{value:1},backgroundRotation:{value:new ot}},vertexShader:st.backgroundCube_vert,fragmentShader:st.backgroundCube_frag},cube:{uniforms:{tCube:{value:null},tFlip:{value:-1},opacity:{value:1}},vertexShader:st.cube_vert,fragmentShader:st.cube_frag},equirect:{uniforms:{tEquirect:{value:null}},vertexShader:st.equirect_vert,fragmentShader:st.equirect_frag},distanceRGBA:{uniforms:Sn([be.common,be.displacementmap,{referencePosition:{value:new Q},nearDistance:{value:1},farDistance:{value:1e3}}]),vertexShader:st.distanceRGBA_vert,fragmentShader:st.distanceRGBA_frag},shadow:{uniforms:Sn([be.lights,be.fog,{color:{value:new vt(0)},opacity:{value:1}}]),vertexShader:st.shadow_vert,fragmentShader:st.shadow_frag}};vi.physical={uniforms:Sn([vi.standard.uniforms,{clearcoat:{value:0},clearcoatMap:{value:null},clearcoatMapTransform:{value:new ot},clearcoatNormalMap:{value:null},clearcoatNormalMapTransform:{value:new ot},clearcoatNormalScale:{value:new rt(1,1)},clearcoatRoughness:{value:0},clearcoatRoughnessMap:{value:null},clearcoatRoughnessMapTransform:{value:new ot},dispersion:{value:0},iridescence:{value:0},iridescenceMap:{value:null},iridescenceMapTransform:{value:new ot},iridescenceIOR:{value:1.3},iridescenceThicknessMinimum:{value:100},iridescenceThicknessMaximum:{value:400},iridescenceThicknessMap:{value:null},iridescenceThicknessMapTransform:{value:new ot},sheen:{value:0},sheenColor:{value:new vt(0)},sheenColorMap:{value:null},sheenColorMapTransform:{value:new ot},sheenRoughness:{value:1},sheenRoughnessMap:{value:null},sheenRoughnessMapTransform:{value:new ot},transmission:{value:0},transmissionMap:{value:null},transmissionMapTransform:{value:new ot},transmissionSamplerSize:{value:new rt},transmissionSamplerMap:{value:null},thickness:{value:0},thicknessMap:{value:null},thicknessMapTransform:{value:new ot},attenuationDistance:{value:0},attenuationColor:{value:new vt(0)},specularColor:{value:new vt(1,1,1)},specularColorMap:{value:null},specularColorMapTransform:{value:new ot},specularIntensity:{value:1},specularIntensityMap:{value:null},specularIntensityMapTransform:{value:new ot},anisotropyVector:{value:new rt},anisotropyMap:{value:null},anisotropyMapTransform:{value:new ot}}]),vertexShader:st.meshphysical_vert,fragmentShader:st.meshphysical_frag};const Nl={r:0,b:0,g:0},qr=new Si,dE=new Gt;function hE(r,e,t,s,a,l,u){const f=new vt(0);let h=l===!0?0:1,p,m,_=null,x=0,S=null;function T(P){let b=P.isScene===!0?P.background:null;return b&&b.isTexture&&(b=(P.backgroundBlurriness>0?t:e).get(b)),b}function w(P){let b=!1;const D=T(P);D===null?y(f,h):D&&D.isColor&&(y(D,1),b=!0);const V=r.xr.getEnvironmentBlendMode();V==="additive"?s.buffers.color.setClear(0,0,0,1,u):V==="alpha-blend"&&s.buffers.color.setClear(0,0,0,0,u),(r.autoClear||b)&&(s.buffers.depth.setTest(!0),s.buffers.depth.setMask(!0),s.buffers.color.setMask(!0),r.clear(r.autoClearColor,r.autoClearDepth,r.autoClearStencil))}function v(P,b){const D=T(b);D&&(D.isCubeTexture||D.mapping===tc)?(m===void 0&&(m=new xi(new la(1,1,1),new Cr({name:"BackgroundCubeMaterial",uniforms:so(vi.backgroundCube.uniforms),vertexShader:vi.backgroundCube.vertexShader,fragmentShader:vi.backgroundCube.fragmentShader,side:Ln,depthTest:!1,depthWrite:!1,fog:!1})),m.geometry.deleteAttribute("normal"),m.geometry.deleteAttribute("uv"),m.onBeforeRender=function(V,O,U){this.matrixWorld.copyPosition(U.matrixWorld)},Object.defineProperty(m.material,"envMap",{get:function(){return this.uniforms.envMap.value}}),a.update(m)),qr.copy(b.backgroundRotation),qr.x*=-1,qr.y*=-1,qr.z*=-1,D.isCubeTexture&&D.isRenderTargetTexture===!1&&(qr.y*=-1,qr.z*=-1),m.material.uniforms.envMap.value=D,m.material.uniforms.flipEnvMap.value=D.isCubeTexture&&D.isRenderTargetTexture===!1?-1:1,m.material.uniforms.backgroundBlurriness.value=b.backgroundBlurriness,m.material.uniforms.backgroundIntensity.value=b.backgroundIntensity,m.material.uniforms.backgroundRotation.value.setFromMatrix4(dE.makeRotationFromEuler(qr)),m.material.toneMapped=Tt.getTransfer(D.colorSpace)!==Ut,(_!==D||x!==D.version||S!==r.toneMapping)&&(m.material.needsUpdate=!0,_=D,x=D.version,S=r.toneMapping),m.layers.enableAll(),P.unshift(m,m.geometry,m.material,0,0,null)):D&&D.isTexture&&(p===void 0&&(p=new xi(new ic(2,2),new Cr({name:"BackgroundMaterial",uniforms:so(vi.background.uniforms),vertexShader:vi.background.vertexShader,fragmentShader:vi.background.fragmentShader,side:Ar,depthTest:!1,depthWrite:!1,fog:!1})),p.geometry.deleteAttribute("normal"),Object.defineProperty(p.material,"map",{get:function(){return this.uniforms.t2D.value}}),a.update(p)),p.material.uniforms.t2D.value=D,p.material.uniforms.backgroundIntensity.value=b.backgroundIntensity,p.material.toneMapped=Tt.getTransfer(D.colorSpace)!==Ut,D.matrixAutoUpdate===!0&&D.updateMatrix(),p.material.uniforms.uvTransform.value.copy(D.matrix),(_!==D||x!==D.version||S!==r.toneMapping)&&(p.material.needsUpdate=!0,_=D,x=D.version,S=r.toneMapping),p.layers.enableAll(),P.unshift(p,p.geometry,p.material,0,0,null))}function y(P,b){P.getRGB(Nl,n_(r)),s.buffers.color.setClear(Nl.r,Nl.g,Nl.b,b,u)}return{getClearColor:function(){return f},setClearColor:function(P,b=1){f.set(P),h=b,y(f,h)},getClearAlpha:function(){return h},setClearAlpha:function(P){h=P,y(f,h)},render:w,addToRenderList:v}}function pE(r,e){const t=r.getParameter(r.MAX_VERTEX_ATTRIBS),s={},a=x(null);let l=a,u=!1;function f(E,C,re,ee,ae){let ue=!1;const Z=_(ee,re,C);l!==Z&&(l=Z,p(l.object)),ue=S(E,ee,re,ae),ue&&T(E,ee,re,ae),ae!==null&&e.update(ae,r.ELEMENT_ARRAY_BUFFER),(ue||u)&&(u=!1,D(E,C,re,ee),ae!==null&&r.bindBuffer(r.ELEMENT_ARRAY_BUFFER,e.get(ae).buffer))}function h(){return r.createVertexArray()}function p(E){return r.bindVertexArray(E)}function m(E){return r.deleteVertexArray(E)}function _(E,C,re){const ee=re.wireframe===!0;let ae=s[E.id];ae===void 0&&(ae={},s[E.id]=ae);let ue=ae[C.id];ue===void 0&&(ue={},ae[C.id]=ue);let Z=ue[ee];return Z===void 0&&(Z=x(h()),ue[ee]=Z),Z}function x(E){const C=[],re=[],ee=[];for(let ae=0;ae<t;ae++)C[ae]=0,re[ae]=0,ee[ae]=0;return{geometry:null,program:null,wireframe:!1,newAttributes:C,enabledAttributes:re,attributeDivisors:ee,object:E,attributes:{},index:null}}function S(E,C,re,ee){const ae=l.attributes,ue=C.attributes;let Z=0;const le=re.getAttributes();for(const F in le)if(le[F].location>=0){const L=ae[F];let X=ue[F];if(X===void 0&&(F==="instanceMatrix"&&E.instanceMatrix&&(X=E.instanceMatrix),F==="instanceColor"&&E.instanceColor&&(X=E.instanceColor)),L===void 0||L.attribute!==X||X&&L.data!==X.data)return!0;Z++}return l.attributesNum!==Z||l.index!==ee}function T(E,C,re,ee){const ae={},ue=C.attributes;let Z=0;const le=re.getAttributes();for(const F in le)if(le[F].location>=0){let L=ue[F];L===void 0&&(F==="instanceMatrix"&&E.instanceMatrix&&(L=E.instanceMatrix),F==="instanceColor"&&E.instanceColor&&(L=E.instanceColor));const X={};X.attribute=L,L&&L.data&&(X.data=L.data),ae[F]=X,Z++}l.attributes=ae,l.attributesNum=Z,l.index=ee}function w(){const E=l.newAttributes;for(let C=0,re=E.length;C<re;C++)E[C]=0}function v(E){y(E,0)}function y(E,C){const re=l.newAttributes,ee=l.enabledAttributes,ae=l.attributeDivisors;re[E]=1,ee[E]===0&&(r.enableVertexAttribArray(E),ee[E]=1),ae[E]!==C&&(r.vertexAttribDivisor(E,C),ae[E]=C)}function P(){const E=l.newAttributes,C=l.enabledAttributes;for(let re=0,ee=C.length;re<ee;re++)C[re]!==E[re]&&(r.disableVertexAttribArray(re),C[re]=0)}function b(E,C,re,ee,ae,ue,Z){Z===!0?r.vertexAttribIPointer(E,C,re,ae,ue):r.vertexAttribPointer(E,C,re,ee,ae,ue)}function D(E,C,re,ee){w();const ae=ee.attributes,ue=re.getAttributes(),Z=C.defaultAttributeValues;for(const le in ue){const F=ue[le];if(F.location>=0){let se=ae[le];if(se===void 0&&(le==="instanceMatrix"&&E.instanceMatrix&&(se=E.instanceMatrix),le==="instanceColor"&&E.instanceColor&&(se=E.instanceColor)),se!==void 0){const L=se.normalized,X=se.itemSize,ve=e.get(se);if(ve===void 0)continue;const Ne=ve.buffer,J=ve.type,fe=ve.bytesPerElement,Se=J===r.INT||J===r.UNSIGNED_INT||se.gpuType===_d;if(se.isInterleavedBufferAttribute){const Me=se.data,Pe=Me.stride,Ge=se.offset;if(Me.isInstancedInterleavedBuffer){for(let dt=0;dt<F.locationSize;dt++)y(F.location+dt,Me.meshPerAttribute);E.isInstancedMesh!==!0&&ee._maxInstanceCount===void 0&&(ee._maxInstanceCount=Me.meshPerAttribute*Me.count)}else for(let dt=0;dt<F.locationSize;dt++)v(F.location+dt);r.bindBuffer(r.ARRAY_BUFFER,Ne);for(let dt=0;dt<F.locationSize;dt++)b(F.location+dt,X/F.locationSize,J,L,Pe*fe,(Ge+X/F.locationSize*dt)*fe,Se)}else{if(se.isInstancedBufferAttribute){for(let Me=0;Me<F.locationSize;Me++)y(F.location+Me,se.meshPerAttribute);E.isInstancedMesh!==!0&&ee._maxInstanceCount===void 0&&(ee._maxInstanceCount=se.meshPerAttribute*se.count)}else for(let Me=0;Me<F.locationSize;Me++)v(F.location+Me);r.bindBuffer(r.ARRAY_BUFFER,Ne);for(let Me=0;Me<F.locationSize;Me++)b(F.location+Me,X/F.locationSize,J,L,X*fe,X/F.locationSize*Me*fe,Se)}}else if(Z!==void 0){const L=Z[le];if(L!==void 0)switch(L.length){case 2:r.vertexAttrib2fv(F.location,L);break;case 3:r.vertexAttrib3fv(F.location,L);break;case 4:r.vertexAttrib4fv(F.location,L);break;default:r.vertexAttrib1fv(F.location,L)}}}}P()}function V(){Y();for(const E in s){const C=s[E];for(const re in C){const ee=C[re];for(const ae in ee)m(ee[ae].object),delete ee[ae];delete C[re]}delete s[E]}}function O(E){if(s[E.id]===void 0)return;const C=s[E.id];for(const re in C){const ee=C[re];for(const ae in ee)m(ee[ae].object),delete ee[ae];delete C[re]}delete s[E.id]}function U(E){for(const C in s){const re=s[C];if(re[E.id]===void 0)continue;const ee=re[E.id];for(const ae in ee)m(ee[ae].object),delete ee[ae];delete re[E.id]}}function Y(){ce(),u=!0,l!==a&&(l=a,p(l.object))}function ce(){a.geometry=null,a.program=null,a.wireframe=!1}return{setup:f,reset:Y,resetDefaultState:ce,dispose:V,releaseStatesOfGeometry:O,releaseStatesOfProgram:U,initAttributes:w,enableAttribute:v,disableUnusedAttributes:P}}function mE(r,e,t){let s;function a(p){s=p}function l(p,m){r.drawArrays(s,p,m),t.update(m,s,1)}function u(p,m,_){_!==0&&(r.drawArraysInstanced(s,p,m,_),t.update(m,s,_))}function f(p,m,_){if(_===0)return;e.get("WEBGL_multi_draw").multiDrawArraysWEBGL(s,p,0,m,0,_);let S=0;for(let T=0;T<_;T++)S+=m[T];t.update(S,s,1)}function h(p,m,_,x){if(_===0)return;const S=e.get("WEBGL_multi_draw");if(S===null)for(let T=0;T<p.length;T++)u(p[T],m[T],x[T]);else{S.multiDrawArraysInstancedWEBGL(s,p,0,m,0,x,0,_);let T=0;for(let w=0;w<_;w++)T+=m[w];for(let w=0;w<x.length;w++)t.update(T,s,x[w])}}this.setMode=a,this.render=l,this.renderInstances=u,this.renderMultiDraw=f,this.renderMultiDrawInstances=h}function gE(r,e,t,s){let a;function l(){if(a!==void 0)return a;if(e.has("EXT_texture_filter_anisotropic")===!0){const U=e.get("EXT_texture_filter_anisotropic");a=r.getParameter(U.MAX_TEXTURE_MAX_ANISOTROPY_EXT)}else a=0;return a}function u(U){return!(U!==di&&s.convert(U)!==r.getParameter(r.IMPLEMENTATION_COLOR_READ_FORMAT))}function f(U){const Y=U===sa&&(e.has("EXT_color_buffer_half_float")||e.has("EXT_color_buffer_float"));return!(U!==Wi&&s.convert(U)!==r.getParameter(r.IMPLEMENTATION_COLOR_READ_TYPE)&&U!==Hi&&!Y)}function h(U){if(U==="highp"){if(r.getShaderPrecisionFormat(r.VERTEX_SHADER,r.HIGH_FLOAT).precision>0&&r.getShaderPrecisionFormat(r.FRAGMENT_SHADER,r.HIGH_FLOAT).precision>0)return"highp";U="mediump"}return U==="mediump"&&r.getShaderPrecisionFormat(r.VERTEX_SHADER,r.MEDIUM_FLOAT).precision>0&&r.getShaderPrecisionFormat(r.FRAGMENT_SHADER,r.MEDIUM_FLOAT).precision>0?"mediump":"lowp"}let p=t.precision!==void 0?t.precision:"highp";const m=h(p);m!==p&&(console.warn("THREE.WebGLRenderer:",p,"not supported, using",m,"instead."),p=m);const _=t.logarithmicDepthBuffer===!0,x=t.reverseDepthBuffer===!0&&e.has("EXT_clip_control");if(x===!0){const U=e.get("EXT_clip_control");U.clipControlEXT(U.LOWER_LEFT_EXT,U.ZERO_TO_ONE_EXT)}const S=r.getParameter(r.MAX_TEXTURE_IMAGE_UNITS),T=r.getParameter(r.MAX_VERTEX_TEXTURE_IMAGE_UNITS),w=r.getParameter(r.MAX_TEXTURE_SIZE),v=r.getParameter(r.MAX_CUBE_MAP_TEXTURE_SIZE),y=r.getParameter(r.MAX_VERTEX_ATTRIBS),P=r.getParameter(r.MAX_VERTEX_UNIFORM_VECTORS),b=r.getParameter(r.MAX_VARYING_VECTORS),D=r.getParameter(r.MAX_FRAGMENT_UNIFORM_VECTORS),V=T>0,O=r.getParameter(r.MAX_SAMPLES);return{isWebGL2:!0,getMaxAnisotropy:l,getMaxPrecision:h,textureFormatReadable:u,textureTypeReadable:f,precision:p,logarithmicDepthBuffer:_,reverseDepthBuffer:x,maxTextures:S,maxVertexTextures:T,maxTextureSize:w,maxCubemapSize:v,maxAttributes:y,maxVertexUniforms:P,maxVaryings:b,maxFragmentUniforms:D,vertexTextures:V,maxSamples:O}}function _E(r){const e=this;let t=null,s=0,a=!1,l=!1;const u=new yr,f=new ot,h={value:null,needsUpdate:!1};this.uniform=h,this.numPlanes=0,this.numIntersection=0,this.init=function(_,x){const S=_.length!==0||x||s!==0||a;return a=x,s=_.length,S},this.beginShadows=function(){l=!0,m(null)},this.endShadows=function(){l=!1},this.setGlobalState=function(_,x){t=m(_,x,0)},this.setState=function(_,x,S){const T=_.clippingPlanes,w=_.clipIntersection,v=_.clipShadows,y=r.get(_);if(!a||T===null||T.length===0||l&&!v)l?m(null):p();else{const P=l?0:s,b=P*4;let D=y.clippingState||null;h.value=D,D=m(T,x,b,S);for(let V=0;V!==b;++V)D[V]=t[V];y.clippingState=D,this.numIntersection=w?this.numPlanes:0,this.numPlanes+=P}};function p(){h.value!==t&&(h.value=t,h.needsUpdate=s>0),e.numPlanes=s,e.numIntersection=0}function m(_,x,S,T){const w=_!==null?_.length:0;let v=null;if(w!==0){if(v=h.value,T!==!0||v===null){const y=S+w*4,P=x.matrixWorldInverse;f.getNormalMatrix(P),(v===null||v.length<y)&&(v=new Float32Array(y));for(let b=0,D=S;b!==w;++b,D+=4)u.copy(_[b]).applyMatrix4(P,f),u.normal.toArray(v,D),v[D+3]=u.constant}h.value=v,h.needsUpdate=!0}return e.numPlanes=w,e.numIntersection=0,v}}function vE(r){let e=new WeakMap;function t(u,f){return f===If?u.mapping=to:f===Uf&&(u.mapping=no),u}function s(u){if(u&&u.isTexture){const f=u.mapping;if(f===If||f===Uf)if(e.has(u)){const h=e.get(u).texture;return t(h,u.mapping)}else{const h=u.image;if(h&&h.height>0){const p=new by(h.height);return p.fromEquirectangularTexture(r,u),e.set(u,p),u.addEventListener("dispose",a),t(p.texture,u.mapping)}else return null}}return u}function a(u){const f=u.target;f.removeEventListener("dispose",a);const h=e.get(f);h!==void 0&&(e.delete(f),h.dispose())}function l(){e=new WeakMap}return{get:s,dispose:l}}class o_ extends i_{constructor(e=-1,t=1,s=1,a=-1,l=.1,u=2e3){super(),this.isOrthographicCamera=!0,this.type="OrthographicCamera",this.zoom=1,this.view=null,this.left=e,this.right=t,this.top=s,this.bottom=a,this.near=l,this.far=u,this.updateProjectionMatrix()}copy(e,t){return super.copy(e,t),this.left=e.left,this.right=e.right,this.top=e.top,this.bottom=e.bottom,this.near=e.near,this.far=e.far,this.zoom=e.zoom,this.view=e.view===null?null:Object.assign({},e.view),this}setViewOffset(e,t,s,a,l,u){this.view===null&&(this.view={enabled:!0,fullWidth:1,fullHeight:1,offsetX:0,offsetY:0,width:1,height:1}),this.view.enabled=!0,this.view.fullWidth=e,this.view.fullHeight=t,this.view.offsetX=s,this.view.offsetY=a,this.view.width=l,this.view.height=u,this.updateProjectionMatrix()}clearViewOffset(){this.view!==null&&(this.view.enabled=!1),this.updateProjectionMatrix()}updateProjectionMatrix(){const e=(this.right-this.left)/(2*this.zoom),t=(this.top-this.bottom)/(2*this.zoom),s=(this.right+this.left)/2,a=(this.top+this.bottom)/2;let l=s-e,u=s+e,f=a+t,h=a-t;if(this.view!==null&&this.view.enabled){const p=(this.right-this.left)/this.view.fullWidth/this.zoom,m=(this.top-this.bottom)/this.view.fullHeight/this.zoom;l+=p*this.view.offsetX,u=l+p*this.view.width,f-=m*this.view.offsetY,h=f-m*this.view.height}this.projectionMatrix.makeOrthographic(l,u,f,h,this.near,this.far,this.coordinateSystem),this.projectionMatrixInverse.copy(this.projectionMatrix).invert()}toJSON(e){const t=super.toJSON(e);return t.object.zoom=this.zoom,t.object.left=this.left,t.object.right=this.right,t.object.top=this.top,t.object.bottom=this.bottom,t.object.near=this.near,t.object.far=this.far,this.view!==null&&(t.object.view=Object.assign({},this.view)),t}}const $s=4,zm=[.125,.215,.35,.446,.526,.582],Jr=20,df=new o_,Hm=new vt;let hf=null,pf=0,mf=0,gf=!1;const Zr=(1+Math.sqrt(5))/2,Ys=1/Zr,Vm=[new Q(-Zr,Ys,0),new Q(Zr,Ys,0),new Q(-Ys,0,Zr),new Q(Ys,0,Zr),new Q(0,Zr,-Ys),new Q(0,Zr,Ys),new Q(-1,1,-1),new Q(1,1,-1),new Q(-1,1,1),new Q(1,1,1)];class Gm{constructor(e){this._renderer=e,this._pingPongRenderTarget=null,this._lodMax=0,this._cubeSize=0,this._lodPlanes=[],this._sizeLods=[],this._sigmas=[],this._blurMaterial=null,this._cubemapMaterial=null,this._equirectMaterial=null,this._compileMaterial(this._blurMaterial)}fromScene(e,t=0,s=.1,a=100){hf=this._renderer.getRenderTarget(),pf=this._renderer.getActiveCubeFace(),mf=this._renderer.getActiveMipmapLevel(),gf=this._renderer.xr.enabled,this._renderer.xr.enabled=!1,this._setSize(256);const l=this._allocateTargets();return l.depthBuffer=!0,this._sceneToCubeUV(e,s,a,l),t>0&&this._blur(l,0,0,t),this._applyPMREM(l),this._cleanup(l),l}fromEquirectangular(e,t=null){return this._fromTexture(e,t)}fromCubemap(e,t=null){return this._fromTexture(e,t)}compileCubemapShader(){this._cubemapMaterial===null&&(this._cubemapMaterial=Xm(),this._compileMaterial(this._cubemapMaterial))}compileEquirectangularShader(){this._equirectMaterial===null&&(this._equirectMaterial=jm(),this._compileMaterial(this._equirectMaterial))}dispose(){this._dispose(),this._cubemapMaterial!==null&&this._cubemapMaterial.dispose(),this._equirectMaterial!==null&&this._equirectMaterial.dispose()}_setSize(e){this._lodMax=Math.floor(Math.log2(e)),this._cubeSize=Math.pow(2,this._lodMax)}_dispose(){this._blurMaterial!==null&&this._blurMaterial.dispose(),this._pingPongRenderTarget!==null&&this._pingPongRenderTarget.dispose();for(let e=0;e<this._lodPlanes.length;e++)this._lodPlanes[e].dispose()}_cleanup(e){this._renderer.setRenderTarget(hf,pf,mf),this._renderer.xr.enabled=gf,e.scissorTest=!1,Il(e,0,0,e.width,e.height)}_fromTexture(e,t){e.mapping===to||e.mapping===no?this._setSize(e.image.length===0?16:e.image[0].width||e.image[0].image.width):this._setSize(e.image.width/4),hf=this._renderer.getRenderTarget(),pf=this._renderer.getActiveCubeFace(),mf=this._renderer.getActiveMipmapLevel(),gf=this._renderer.xr.enabled,this._renderer.xr.enabled=!1;const s=t||this._allocateTargets();return this._textureToCubeUV(e,s),this._applyPMREM(s),this._cleanup(s),s}_allocateTargets(){const e=3*Math.max(this._cubeSize,112),t=4*this._cubeSize,s={magFilter:ui,minFilter:ui,generateMipmaps:!1,type:sa,format:di,colorSpace:br,depthBuffer:!1},a=Wm(e,t,s);if(this._pingPongRenderTarget===null||this._pingPongRenderTarget.width!==e||this._pingPongRenderTarget.height!==t){this._pingPongRenderTarget!==null&&this._dispose(),this._pingPongRenderTarget=Wm(e,t,s);const{_lodMax:l}=this;({sizeLods:this._sizeLods,lodPlanes:this._lodPlanes,sigmas:this._sigmas}=xE(l)),this._blurMaterial=yE(l,e,t)}return a}_compileMaterial(e){const t=new xi(this._lodPlanes[0],e);this._renderer.compile(t,df)}_sceneToCubeUV(e,t,s,a){const f=new Zn(90,1,t,s),h=[1,-1,1,1,1,1],p=[1,1,1,-1,-1,-1],m=this._renderer,_=m.autoClear,x=m.toneMapping;m.getClearColor(Hm),m.toneMapping=wr,m.autoClear=!1;const S=new Jg({name:"PMREM.Background",side:Ln,depthWrite:!1,depthTest:!1}),T=new xi(new la,S);let w=!1;const v=e.background;v?v.isColor&&(S.color.copy(v),e.background=null,w=!0):(S.color.copy(Hm),w=!0);for(let y=0;y<6;y++){const P=y%3;P===0?(f.up.set(0,h[y],0),f.lookAt(p[y],0,0)):P===1?(f.up.set(0,0,h[y]),f.lookAt(0,p[y],0)):(f.up.set(0,h[y],0),f.lookAt(0,0,p[y]));const b=this._cubeSize;Il(a,P*b,y>2?b:0,b,b),m.setRenderTarget(a),w&&m.render(T,f),m.render(e,f)}T.geometry.dispose(),T.material.dispose(),m.toneMapping=x,m.autoClear=_,e.background=v}_textureToCubeUV(e,t){const s=this._renderer,a=e.mapping===to||e.mapping===no;a?(this._cubemapMaterial===null&&(this._cubemapMaterial=Xm()),this._cubemapMaterial.uniforms.flipEnvMap.value=e.isRenderTargetTexture===!1?-1:1):this._equirectMaterial===null&&(this._equirectMaterial=jm());const l=a?this._cubemapMaterial:this._equirectMaterial,u=new xi(this._lodPlanes[0],l),f=l.uniforms;f.envMap.value=e;const h=this._cubeSize;Il(t,0,0,3*h,2*h),s.setRenderTarget(t),s.render(u,df)}_applyPMREM(e){const t=this._renderer,s=t.autoClear;t.autoClear=!1;const a=this._lodPlanes.length;for(let l=1;l<a;l++){const u=Math.sqrt(this._sigmas[l]*this._sigmas[l]-this._sigmas[l-1]*this._sigmas[l-1]),f=Vm[(a-l-1)%Vm.length];this._blur(e,l-1,l,u,f)}t.autoClear=s}_blur(e,t,s,a,l){const u=this._pingPongRenderTarget;this._halfBlur(e,u,t,s,a,"latitudinal",l),this._halfBlur(u,e,s,s,a,"longitudinal",l)}_halfBlur(e,t,s,a,l,u,f){const h=this._renderer,p=this._blurMaterial;u!=="latitudinal"&&u!=="longitudinal"&&console.error("blur direction must be either latitudinal or longitudinal!");const m=3,_=new xi(this._lodPlanes[a],p),x=p.uniforms,S=this._sizeLods[s]-1,T=isFinite(l)?Math.PI/(2*S):2*Math.PI/(2*Jr-1),w=l/T,v=isFinite(l)?1+Math.floor(m*w):Jr;v>Jr&&console.warn(`sigmaRadians, ${l}, is too large and will clip, as it requested ${v} samples when the maximum is set to ${Jr}`);const y=[];let P=0;for(let U=0;U<Jr;++U){const Y=U/w,ce=Math.exp(-Y*Y/2);y.push(ce),U===0?P+=ce:U<v&&(P+=2*ce)}for(let U=0;U<y.length;U++)y[U]=y[U]/P;x.envMap.value=e.texture,x.samples.value=v,x.weights.value=y,x.latitudinal.value=u==="latitudinal",f&&(x.poleAxis.value=f);const{_lodMax:b}=this;x.dTheta.value=T,x.mipInt.value=b-s;const D=this._sizeLods[a],V=3*D*(a>b-$s?a-b+$s:0),O=4*(this._cubeSize-D);Il(t,V,O,3*D,2*D),h.setRenderTarget(t),h.render(_,df)}}function xE(r){const e=[],t=[],s=[];let a=r;const l=r-$s+1+zm.length;for(let u=0;u<l;u++){const f=Math.pow(2,a);t.push(f);let h=1/f;u>r-$s?h=zm[u-r+$s-1]:u===0&&(h=0),s.push(h);const p=1/(f-2),m=-p,_=1+p,x=[m,m,_,m,_,_,m,m,_,_,m,_],S=6,T=6,w=3,v=2,y=1,P=new Float32Array(w*T*S),b=new Float32Array(v*T*S),D=new Float32Array(y*T*S);for(let O=0;O<S;O++){const U=O%3*2/3-1,Y=O>2?0:-1,ce=[U,Y,0,U+2/3,Y,0,U+2/3,Y+1,0,U,Y,0,U+2/3,Y+1,0,U,Y+1,0];P.set(ce,w*T*O),b.set(x,v*T*O);const E=[O,O,O,O,O,O];D.set(E,y*T*O)}const V=new ji;V.setAttribute("position",new Vn(P,w)),V.setAttribute("uv",new Vn(b,v)),V.setAttribute("faceIndex",new Vn(D,y)),e.push(V),a>$s&&a--}return{lodPlanes:e,sizeLods:t,sigmas:s}}function Wm(r,e,t){const s=new is(r,e,t);return s.texture.mapping=tc,s.texture.name="PMREM.cubeUv",s.scissorTest=!0,s}function Il(r,e,t,s,a){r.viewport.set(e,t,s,a),r.scissor.set(e,t,s,a)}function yE(r,e,t){const s=new Float32Array(Jr),a=new Q(0,1,0);return new Cr({name:"SphericalGaussianBlur",defines:{n:Jr,CUBEUV_TEXEL_WIDTH:1/e,CUBEUV_TEXEL_HEIGHT:1/t,CUBEUV_MAX_MIP:`${r}.0`},uniforms:{envMap:{value:null},samples:{value:1},weights:{value:s},latitudinal:{value:!1},dTheta:{value:0},mipInt:{value:0},poleAxis:{value:a}},vertexShader:Ad(),fragmentShader:`

			precision mediump float;
			precision mediump int;

			varying vec3 vOutputDirection;

			uniform sampler2D envMap;
			uniform int samples;
			uniform float weights[ n ];
			uniform bool latitudinal;
			uniform float dTheta;
			uniform float mipInt;
			uniform vec3 poleAxis;

			#define ENVMAP_TYPE_CUBE_UV
			#include <cube_uv_reflection_fragment>

			vec3 getSample( float theta, vec3 axis ) {

				float cosTheta = cos( theta );
				// Rodrigues' axis-angle rotation
				vec3 sampleDirection = vOutputDirection * cosTheta
					+ cross( axis, vOutputDirection ) * sin( theta )
					+ axis * dot( axis, vOutputDirection ) * ( 1.0 - cosTheta );

				return bilinearCubeUV( envMap, sampleDirection, mipInt );

			}

			void main() {

				vec3 axis = latitudinal ? poleAxis : cross( poleAxis, vOutputDirection );

				if ( all( equal( axis, vec3( 0.0 ) ) ) ) {

					axis = vec3( vOutputDirection.z, 0.0, - vOutputDirection.x );

				}

				axis = normalize( axis );

				gl_FragColor = vec4( 0.0, 0.0, 0.0, 1.0 );
				gl_FragColor.rgb += weights[ 0 ] * getSample( 0.0, axis );

				for ( int i = 1; i < n; i++ ) {

					if ( i >= samples ) {

						break;

					}

					float theta = dTheta * float( i );
					gl_FragColor.rgb += weights[ i ] * getSample( -1.0 * theta, axis );
					gl_FragColor.rgb += weights[ i ] * getSample( theta, axis );

				}

			}
		`,blending:Tr,depthTest:!1,depthWrite:!1})}function jm(){return new Cr({name:"EquirectangularToCubeUV",uniforms:{envMap:{value:null}},vertexShader:Ad(),fragmentShader:`

			precision mediump float;
			precision mediump int;

			varying vec3 vOutputDirection;

			uniform sampler2D envMap;

			#include <common>

			void main() {

				vec3 outputDirection = normalize( vOutputDirection );
				vec2 uv = equirectUv( outputDirection );

				gl_FragColor = vec4( texture2D ( envMap, uv ).rgb, 1.0 );

			}
		`,blending:Tr,depthTest:!1,depthWrite:!1})}function Xm(){return new Cr({name:"CubemapToCubeUV",uniforms:{envMap:{value:null},flipEnvMap:{value:-1}},vertexShader:Ad(),fragmentShader:`

			precision mediump float;
			precision mediump int;

			uniform float flipEnvMap;

			varying vec3 vOutputDirection;

			uniform samplerCube envMap;

			void main() {

				gl_FragColor = textureCube( envMap, vec3( flipEnvMap * vOutputDirection.x, vOutputDirection.yz ) );

			}
		`,blending:Tr,depthTest:!1,depthWrite:!1})}function Ad(){return`

		precision mediump float;
		precision mediump int;

		attribute float faceIndex;

		varying vec3 vOutputDirection;

		// RH coordinate system; PMREM face-indexing convention
		vec3 getDirection( vec2 uv, float face ) {

			uv = 2.0 * uv - 1.0;

			vec3 direction = vec3( uv, 1.0 );

			if ( face == 0.0 ) {

				direction = direction.zyx; // ( 1, v, u ) pos x

			} else if ( face == 1.0 ) {

				direction = direction.xzy;
				direction.xz *= -1.0; // ( -u, 1, -v ) pos y

			} else if ( face == 2.0 ) {

				direction.x *= -1.0; // ( -u, v, 1 ) pos z

			} else if ( face == 3.0 ) {

				direction = direction.zyx;
				direction.xz *= -1.0; // ( -1, v, -u ) neg x

			} else if ( face == 4.0 ) {

				direction = direction.xzy;
				direction.xy *= -1.0; // ( -u, -1, v ) neg y

			} else if ( face == 5.0 ) {

				direction.z *= -1.0; // ( u, v, -1 ) neg z

			}

			return direction;

		}

		void main() {

			vOutputDirection = getDirection( uv, faceIndex );
			gl_Position = vec4( position, 1.0 );

		}
	`}function SE(r){let e=new WeakMap,t=null;function s(f){if(f&&f.isTexture){const h=f.mapping,p=h===If||h===Uf,m=h===to||h===no;if(p||m){let _=e.get(f);const x=_!==void 0?_.texture.pmremVersion:0;if(f.isRenderTargetTexture&&f.pmremVersion!==x)return t===null&&(t=new Gm(r)),_=p?t.fromEquirectangular(f,_):t.fromCubemap(f,_),_.texture.pmremVersion=f.pmremVersion,e.set(f,_),_.texture;if(_!==void 0)return _.texture;{const S=f.image;return p&&S&&S.height>0||m&&S&&a(S)?(t===null&&(t=new Gm(r)),_=p?t.fromEquirectangular(f):t.fromCubemap(f),_.texture.pmremVersion=f.pmremVersion,e.set(f,_),f.addEventListener("dispose",l),_.texture):null}}}return f}function a(f){let h=0;const p=6;for(let m=0;m<p;m++)f[m]!==void 0&&h++;return h===p}function l(f){const h=f.target;h.removeEventListener("dispose",l);const p=e.get(h);p!==void 0&&(e.delete(h),p.dispose())}function u(){e=new WeakMap,t!==null&&(t.dispose(),t=null)}return{get:s,dispose:u}}function ME(r){const e={};function t(s){if(e[s]!==void 0)return e[s];let a;switch(s){case"WEBGL_depth_texture":a=r.getExtension("WEBGL_depth_texture")||r.getExtension("MOZ_WEBGL_depth_texture")||r.getExtension("WEBKIT_WEBGL_depth_texture");break;case"EXT_texture_filter_anisotropic":a=r.getExtension("EXT_texture_filter_anisotropic")||r.getExtension("MOZ_EXT_texture_filter_anisotropic")||r.getExtension("WEBKIT_EXT_texture_filter_anisotropic");break;case"WEBGL_compressed_texture_s3tc":a=r.getExtension("WEBGL_compressed_texture_s3tc")||r.getExtension("MOZ_WEBGL_compressed_texture_s3tc")||r.getExtension("WEBKIT_WEBGL_compressed_texture_s3tc");break;case"WEBGL_compressed_texture_pvrtc":a=r.getExtension("WEBGL_compressed_texture_pvrtc")||r.getExtension("WEBKIT_WEBGL_compressed_texture_pvrtc");break;default:a=r.getExtension(s)}return e[s]=a,a}return{has:function(s){return t(s)!==null},init:function(){t("EXT_color_buffer_float"),t("WEBGL_clip_cull_distance"),t("OES_texture_float_linear"),t("EXT_color_buffer_half_float"),t("WEBGL_multisampled_render_to_texture"),t("WEBGL_render_shared_exponent")},get:function(s){const a=t(s);return a===null&&Wl("THREE.WebGLRenderer: "+s+" extension not supported."),a}}}function EE(r,e,t,s){const a={},l=new WeakMap;function u(_){const x=_.target;x.index!==null&&e.remove(x.index);for(const T in x.attributes)e.remove(x.attributes[T]);for(const T in x.morphAttributes){const w=x.morphAttributes[T];for(let v=0,y=w.length;v<y;v++)e.remove(w[v])}x.removeEventListener("dispose",u),delete a[x.id];const S=l.get(x);S&&(e.remove(S),l.delete(x)),s.releaseStatesOfGeometry(x),x.isInstancedBufferGeometry===!0&&delete x._maxInstanceCount,t.memory.geometries--}function f(_,x){return a[x.id]===!0||(x.addEventListener("dispose",u),a[x.id]=!0,t.memory.geometries++),x}function h(_){const x=_.attributes;for(const T in x)e.update(x[T],r.ARRAY_BUFFER);const S=_.morphAttributes;for(const T in S){const w=S[T];for(let v=0,y=w.length;v<y;v++)e.update(w[v],r.ARRAY_BUFFER)}}function p(_){const x=[],S=_.index,T=_.attributes.position;let w=0;if(S!==null){const P=S.array;w=S.version;for(let b=0,D=P.length;b<D;b+=3){const V=P[b+0],O=P[b+1],U=P[b+2];x.push(V,O,O,U,U,V)}}else if(T!==void 0){const P=T.array;w=T.version;for(let b=0,D=P.length/3-1;b<D;b+=3){const V=b+0,O=b+1,U=b+2;x.push(V,O,O,U,U,V)}}else return;const v=new(Yg(x)?t_:e_)(x,1);v.version=w;const y=l.get(_);y&&e.remove(y),l.set(_,v)}function m(_){const x=l.get(_);if(x){const S=_.index;S!==null&&x.version<S.version&&p(_)}else p(_);return l.get(_)}return{get:f,update:h,getWireframeAttribute:m}}function TE(r,e,t){let s;function a(x){s=x}let l,u;function f(x){l=x.type,u=x.bytesPerElement}function h(x,S){r.drawElements(s,S,l,x*u),t.update(S,s,1)}function p(x,S,T){T!==0&&(r.drawElementsInstanced(s,S,l,x*u,T),t.update(S,s,T))}function m(x,S,T){if(T===0)return;e.get("WEBGL_multi_draw").multiDrawElementsWEBGL(s,S,0,l,x,0,T);let v=0;for(let y=0;y<T;y++)v+=S[y];t.update(v,s,1)}function _(x,S,T,w){if(T===0)return;const v=e.get("WEBGL_multi_draw");if(v===null)for(let y=0;y<x.length;y++)p(x[y]/u,S[y],w[y]);else{v.multiDrawElementsInstancedWEBGL(s,S,0,l,x,0,w,0,T);let y=0;for(let P=0;P<T;P++)y+=S[P];for(let P=0;P<w.length;P++)t.update(y,s,w[P])}}this.setMode=a,this.setIndex=f,this.render=h,this.renderInstances=p,this.renderMultiDraw=m,this.renderMultiDrawInstances=_}function wE(r){const e={geometries:0,textures:0},t={frame:0,calls:0,triangles:0,points:0,lines:0};function s(l,u,f){switch(t.calls++,u){case r.TRIANGLES:t.triangles+=f*(l/3);break;case r.LINES:t.lines+=f*(l/2);break;case r.LINE_STRIP:t.lines+=f*(l-1);break;case r.LINE_LOOP:t.lines+=f*l;break;case r.POINTS:t.points+=f*l;break;default:console.error("THREE.WebGLInfo: Unknown draw mode:",u);break}}function a(){t.calls=0,t.triangles=0,t.points=0,t.lines=0}return{memory:e,render:t,programs:null,autoReset:!0,reset:a,update:s}}function AE(r,e,t){const s=new WeakMap,a=new Vt;function l(u,f,h){const p=u.morphTargetInfluences,m=f.morphAttributes.position||f.morphAttributes.normal||f.morphAttributes.color,_=m!==void 0?m.length:0;let x=s.get(f);if(x===void 0||x.count!==_){let E=function(){Y.dispose(),s.delete(f),f.removeEventListener("dispose",E)};var S=E;x!==void 0&&x.texture.dispose();const T=f.morphAttributes.position!==void 0,w=f.morphAttributes.normal!==void 0,v=f.morphAttributes.color!==void 0,y=f.morphAttributes.position||[],P=f.morphAttributes.normal||[],b=f.morphAttributes.color||[];let D=0;T===!0&&(D=1),w===!0&&(D=2),v===!0&&(D=3);let V=f.attributes.position.count*D,O=1;V>e.maxTextureSize&&(O=Math.ceil(V/e.maxTextureSize),V=e.maxTextureSize);const U=new Float32Array(V*O*4*_),Y=new $g(U,V,O,_);Y.type=Hi,Y.needsUpdate=!0;const ce=D*4;for(let C=0;C<_;C++){const re=y[C],ee=P[C],ae=b[C],ue=V*O*4*C;for(let Z=0;Z<re.count;Z++){const le=Z*ce;T===!0&&(a.fromBufferAttribute(re,Z),U[ue+le+0]=a.x,U[ue+le+1]=a.y,U[ue+le+2]=a.z,U[ue+le+3]=0),w===!0&&(a.fromBufferAttribute(ee,Z),U[ue+le+4]=a.x,U[ue+le+5]=a.y,U[ue+le+6]=a.z,U[ue+le+7]=0),v===!0&&(a.fromBufferAttribute(ae,Z),U[ue+le+8]=a.x,U[ue+le+9]=a.y,U[ue+le+10]=a.z,U[ue+le+11]=ae.itemSize===4?a.w:1)}}x={count:_,texture:Y,size:new rt(V,O)},s.set(f,x),f.addEventListener("dispose",E)}if(u.isInstancedMesh===!0&&u.morphTexture!==null)h.getUniforms().setValue(r,"morphTexture",u.morphTexture,t);else{let T=0;for(let v=0;v<p.length;v++)T+=p[v];const w=f.morphTargetsRelative?1:1-T;h.getUniforms().setValue(r,"morphTargetBaseInfluence",w),h.getUniforms().setValue(r,"morphTargetInfluences",p)}h.getUniforms().setValue(r,"morphTargetsTexture",x.texture,t),h.getUniforms().setValue(r,"morphTargetsTextureSize",x.size)}return{update:l}}function CE(r,e,t,s){let a=new WeakMap;function l(h){const p=s.render.frame,m=h.geometry,_=e.get(h,m);if(a.get(_)!==p&&(e.update(_),a.set(_,p)),h.isInstancedMesh&&(h.hasEventListener("dispose",f)===!1&&h.addEventListener("dispose",f),a.get(h)!==p&&(t.update(h.instanceMatrix,r.ARRAY_BUFFER),h.instanceColor!==null&&t.update(h.instanceColor,r.ARRAY_BUFFER),a.set(h,p))),h.isSkinnedMesh){const x=h.skeleton;a.get(x)!==p&&(x.update(),a.set(x,p))}return _}function u(){a=new WeakMap}function f(h){const p=h.target;p.removeEventListener("dispose",f),t.remove(p.instanceMatrix),p.instanceColor!==null&&t.remove(p.instanceColor)}return{update:l,dispose:u}}class a_ extends Dn{constructor(e,t,s,a,l,u,f,h,p,m=Qs){if(m!==Qs&&m!==ro)throw new Error("DepthTexture format must be either THREE.DepthFormat or THREE.DepthStencilFormat");s===void 0&&m===Qs&&(s=ns),s===void 0&&m===ro&&(s=io),super(null,a,l,u,f,h,m,s,p),this.isDepthTexture=!0,this.image={width:e,height:t},this.magFilter=f!==void 0?f:Qn,this.minFilter=h!==void 0?h:Qn,this.flipY=!1,this.generateMipmaps=!1,this.compareFunction=null}copy(e){return super.copy(e),this.compareFunction=e.compareFunction,this}toJSON(e){const t=super.toJSON(e);return this.compareFunction!==null&&(t.compareFunction=this.compareFunction),t}}const l_=new Dn,Ym=new a_(1,1),c_=new $g,u_=new hy,f_=new r_,qm=[],$m=[],Km=new Float32Array(16),Zm=new Float32Array(9),Qm=new Float32Array(4);function lo(r,e,t){const s=r[0];if(s<=0||s>0)return r;const a=e*t;let l=qm[a];if(l===void 0&&(l=new Float32Array(a),qm[a]=l),e!==0){s.toArray(l,0);for(let u=1,f=0;u!==e;++u)f+=t,r[u].toArray(l,f)}return l}function Qt(r,e){if(r.length!==e.length)return!1;for(let t=0,s=r.length;t<s;t++)if(r[t]!==e[t])return!1;return!0}function Jt(r,e){for(let t=0,s=e.length;t<s;t++)r[t]=e[t]}function rc(r,e){let t=$m[e];t===void 0&&(t=new Int32Array(e),$m[e]=t);for(let s=0;s!==e;++s)t[s]=r.allocateTextureUnit();return t}function RE(r,e){const t=this.cache;t[0]!==e&&(r.uniform1f(this.addr,e),t[0]=e)}function bE(r,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y)&&(r.uniform2f(this.addr,e.x,e.y),t[0]=e.x,t[1]=e.y);else{if(Qt(t,e))return;r.uniform2fv(this.addr,e),Jt(t,e)}}function PE(r,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y||t[2]!==e.z)&&(r.uniform3f(this.addr,e.x,e.y,e.z),t[0]=e.x,t[1]=e.y,t[2]=e.z);else if(e.r!==void 0)(t[0]!==e.r||t[1]!==e.g||t[2]!==e.b)&&(r.uniform3f(this.addr,e.r,e.g,e.b),t[0]=e.r,t[1]=e.g,t[2]=e.b);else{if(Qt(t,e))return;r.uniform3fv(this.addr,e),Jt(t,e)}}function LE(r,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y||t[2]!==e.z||t[3]!==e.w)&&(r.uniform4f(this.addr,e.x,e.y,e.z,e.w),t[0]=e.x,t[1]=e.y,t[2]=e.z,t[3]=e.w);else{if(Qt(t,e))return;r.uniform4fv(this.addr,e),Jt(t,e)}}function DE(r,e){const t=this.cache,s=e.elements;if(s===void 0){if(Qt(t,e))return;r.uniformMatrix2fv(this.addr,!1,e),Jt(t,e)}else{if(Qt(t,s))return;Qm.set(s),r.uniformMatrix2fv(this.addr,!1,Qm),Jt(t,s)}}function NE(r,e){const t=this.cache,s=e.elements;if(s===void 0){if(Qt(t,e))return;r.uniformMatrix3fv(this.addr,!1,e),Jt(t,e)}else{if(Qt(t,s))return;Zm.set(s),r.uniformMatrix3fv(this.addr,!1,Zm),Jt(t,s)}}function IE(r,e){const t=this.cache,s=e.elements;if(s===void 0){if(Qt(t,e))return;r.uniformMatrix4fv(this.addr,!1,e),Jt(t,e)}else{if(Qt(t,s))return;Km.set(s),r.uniformMatrix4fv(this.addr,!1,Km),Jt(t,s)}}function UE(r,e){const t=this.cache;t[0]!==e&&(r.uniform1i(this.addr,e),t[0]=e)}function FE(r,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y)&&(r.uniform2i(this.addr,e.x,e.y),t[0]=e.x,t[1]=e.y);else{if(Qt(t,e))return;r.uniform2iv(this.addr,e),Jt(t,e)}}function OE(r,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y||t[2]!==e.z)&&(r.uniform3i(this.addr,e.x,e.y,e.z),t[0]=e.x,t[1]=e.y,t[2]=e.z);else{if(Qt(t,e))return;r.uniform3iv(this.addr,e),Jt(t,e)}}function kE(r,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y||t[2]!==e.z||t[3]!==e.w)&&(r.uniform4i(this.addr,e.x,e.y,e.z,e.w),t[0]=e.x,t[1]=e.y,t[2]=e.z,t[3]=e.w);else{if(Qt(t,e))return;r.uniform4iv(this.addr,e),Jt(t,e)}}function BE(r,e){const t=this.cache;t[0]!==e&&(r.uniform1ui(this.addr,e),t[0]=e)}function zE(r,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y)&&(r.uniform2ui(this.addr,e.x,e.y),t[0]=e.x,t[1]=e.y);else{if(Qt(t,e))return;r.uniform2uiv(this.addr,e),Jt(t,e)}}function HE(r,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y||t[2]!==e.z)&&(r.uniform3ui(this.addr,e.x,e.y,e.z),t[0]=e.x,t[1]=e.y,t[2]=e.z);else{if(Qt(t,e))return;r.uniform3uiv(this.addr,e),Jt(t,e)}}function VE(r,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y||t[2]!==e.z||t[3]!==e.w)&&(r.uniform4ui(this.addr,e.x,e.y,e.z,e.w),t[0]=e.x,t[1]=e.y,t[2]=e.z,t[3]=e.w);else{if(Qt(t,e))return;r.uniform4uiv(this.addr,e),Jt(t,e)}}function GE(r,e,t){const s=this.cache,a=t.allocateTextureUnit();s[0]!==a&&(r.uniform1i(this.addr,a),s[0]=a);let l;this.type===r.SAMPLER_2D_SHADOW?(Ym.compareFunction=Xg,l=Ym):l=l_,t.setTexture2D(e||l,a)}function WE(r,e,t){const s=this.cache,a=t.allocateTextureUnit();s[0]!==a&&(r.uniform1i(this.addr,a),s[0]=a),t.setTexture3D(e||u_,a)}function jE(r,e,t){const s=this.cache,a=t.allocateTextureUnit();s[0]!==a&&(r.uniform1i(this.addr,a),s[0]=a),t.setTextureCube(e||f_,a)}function XE(r,e,t){const s=this.cache,a=t.allocateTextureUnit();s[0]!==a&&(r.uniform1i(this.addr,a),s[0]=a),t.setTexture2DArray(e||c_,a)}function YE(r){switch(r){case 5126:return RE;case 35664:return bE;case 35665:return PE;case 35666:return LE;case 35674:return DE;case 35675:return NE;case 35676:return IE;case 5124:case 35670:return UE;case 35667:case 35671:return FE;case 35668:case 35672:return OE;case 35669:case 35673:return kE;case 5125:return BE;case 36294:return zE;case 36295:return HE;case 36296:return VE;case 35678:case 36198:case 36298:case 36306:case 35682:return GE;case 35679:case 36299:case 36307:return WE;case 35680:case 36300:case 36308:case 36293:return jE;case 36289:case 36303:case 36311:case 36292:return XE}}function qE(r,e){r.uniform1fv(this.addr,e)}function $E(r,e){const t=lo(e,this.size,2);r.uniform2fv(this.addr,t)}function KE(r,e){const t=lo(e,this.size,3);r.uniform3fv(this.addr,t)}function ZE(r,e){const t=lo(e,this.size,4);r.uniform4fv(this.addr,t)}function QE(r,e){const t=lo(e,this.size,4);r.uniformMatrix2fv(this.addr,!1,t)}function JE(r,e){const t=lo(e,this.size,9);r.uniformMatrix3fv(this.addr,!1,t)}function eT(r,e){const t=lo(e,this.size,16);r.uniformMatrix4fv(this.addr,!1,t)}function tT(r,e){r.uniform1iv(this.addr,e)}function nT(r,e){r.uniform2iv(this.addr,e)}function iT(r,e){r.uniform3iv(this.addr,e)}function rT(r,e){r.uniform4iv(this.addr,e)}function sT(r,e){r.uniform1uiv(this.addr,e)}function oT(r,e){r.uniform2uiv(this.addr,e)}function aT(r,e){r.uniform3uiv(this.addr,e)}function lT(r,e){r.uniform4uiv(this.addr,e)}function cT(r,e,t){const s=this.cache,a=e.length,l=rc(t,a);Qt(s,l)||(r.uniform1iv(this.addr,l),Jt(s,l));for(let u=0;u!==a;++u)t.setTexture2D(e[u]||l_,l[u])}function uT(r,e,t){const s=this.cache,a=e.length,l=rc(t,a);Qt(s,l)||(r.uniform1iv(this.addr,l),Jt(s,l));for(let u=0;u!==a;++u)t.setTexture3D(e[u]||u_,l[u])}function fT(r,e,t){const s=this.cache,a=e.length,l=rc(t,a);Qt(s,l)||(r.uniform1iv(this.addr,l),Jt(s,l));for(let u=0;u!==a;++u)t.setTextureCube(e[u]||f_,l[u])}function dT(r,e,t){const s=this.cache,a=e.length,l=rc(t,a);Qt(s,l)||(r.uniform1iv(this.addr,l),Jt(s,l));for(let u=0;u!==a;++u)t.setTexture2DArray(e[u]||c_,l[u])}function hT(r){switch(r){case 5126:return qE;case 35664:return $E;case 35665:return KE;case 35666:return ZE;case 35674:return QE;case 35675:return JE;case 35676:return eT;case 5124:case 35670:return tT;case 35667:case 35671:return nT;case 35668:case 35672:return iT;case 35669:case 35673:return rT;case 5125:return sT;case 36294:return oT;case 36295:return aT;case 36296:return lT;case 35678:case 36198:case 36298:case 36306:case 35682:return cT;case 35679:case 36299:case 36307:return uT;case 35680:case 36300:case 36308:case 36293:return fT;case 36289:case 36303:case 36311:case 36292:return dT}}class pT{constructor(e,t,s){this.id=e,this.addr=s,this.cache=[],this.type=t.type,this.setValue=YE(t.type)}}class mT{constructor(e,t,s){this.id=e,this.addr=s,this.cache=[],this.type=t.type,this.size=t.size,this.setValue=hT(t.type)}}class gT{constructor(e){this.id=e,this.seq=[],this.map={}}setValue(e,t,s){const a=this.seq;for(let l=0,u=a.length;l!==u;++l){const f=a[l];f.setValue(e,t[f.id],s)}}}const _f=/(\w+)(\])?(\[|\.)?/g;function Jm(r,e){r.seq.push(e),r.map[e.id]=e}function _T(r,e,t){const s=r.name,a=s.length;for(_f.lastIndex=0;;){const l=_f.exec(s),u=_f.lastIndex;let f=l[1];const h=l[2]==="]",p=l[3];if(h&&(f=f|0),p===void 0||p==="["&&u+2===a){Jm(t,p===void 0?new pT(f,r,e):new mT(f,r,e));break}else{let _=t.map[f];_===void 0&&(_=new gT(f),Jm(t,_)),t=_}}}class jl{constructor(e,t){this.seq=[],this.map={};const s=e.getProgramParameter(t,e.ACTIVE_UNIFORMS);for(let a=0;a<s;++a){const l=e.getActiveUniform(t,a),u=e.getUniformLocation(t,l.name);_T(l,u,this)}}setValue(e,t,s,a){const l=this.map[t];l!==void 0&&l.setValue(e,s,a)}setOptional(e,t,s){const a=t[s];a!==void 0&&this.setValue(e,s,a)}static upload(e,t,s,a){for(let l=0,u=t.length;l!==u;++l){const f=t[l],h=s[f.id];h.needsUpdate!==!1&&f.setValue(e,h.value,a)}}static seqWithValue(e,t){const s=[];for(let a=0,l=e.length;a!==l;++a){const u=e[a];u.id in t&&s.push(u)}return s}}function eg(r,e,t){const s=r.createShader(e);return r.shaderSource(s,t),r.compileShader(s),s}const vT=37297;let xT=0;function yT(r,e){const t=r.split(`
`),s=[],a=Math.max(e-6,0),l=Math.min(e+6,t.length);for(let u=a;u<l;u++){const f=u+1;s.push(`${f===e?">":" "} ${f}: ${t[u]}`)}return s.join(`
`)}function ST(r){const e=Tt.getPrimaries(Tt.workingColorSpace),t=Tt.getPrimaries(r);let s;switch(e===t?s="":e===$l&&t===ql?s="LinearDisplayP3ToLinearSRGB":e===ql&&t===$l&&(s="LinearSRGBToLinearDisplayP3"),r){case br:case nc:return[s,"LinearTransferOETF"];case ci:case Ed:return[s,"sRGBTransferOETF"];default:return console.warn("THREE.WebGLProgram: Unsupported color space:",r),[s,"LinearTransferOETF"]}}function tg(r,e,t){const s=r.getShaderParameter(e,r.COMPILE_STATUS),a=r.getShaderInfoLog(e).trim();if(s&&a==="")return"";const l=/ERROR: 0:(\d+)/.exec(a);if(l){const u=parseInt(l[1]);return t.toUpperCase()+`

`+a+`

`+yT(r.getShaderSource(e),u)}else return a}function MT(r,e){const t=ST(e);return`vec4 ${r}( vec4 value ) { return ${t[0]}( ${t[1]}( value ) ); }`}function ET(r,e){let t;switch(e){case kx:t="Linear";break;case Bx:t="Reinhard";break;case zx:t="Cineon";break;case Hx:t="ACESFilmic";break;case Gx:t="AgX";break;case Wx:t="Neutral";break;case Vx:t="Custom";break;default:console.warn("THREE.WebGLProgram: Unsupported toneMapping:",e),t="Linear"}return"vec3 "+r+"( vec3 color ) { return "+t+"ToneMapping( color ); }"}const Ul=new Q;function TT(){Tt.getLuminanceCoefficients(Ul);const r=Ul.x.toFixed(4),e=Ul.y.toFixed(4),t=Ul.z.toFixed(4);return["float luminance( const in vec3 rgb ) {",`	const vec3 weights = vec3( ${r}, ${e}, ${t} );`,"	return dot( weights, rgb );","}"].join(`
`)}function wT(r){return[r.extensionClipCullDistance?"#extension GL_ANGLE_clip_cull_distance : require":"",r.extensionMultiDraw?"#extension GL_ANGLE_multi_draw : require":""].filter(Qo).join(`
`)}function AT(r){const e=[];for(const t in r){const s=r[t];s!==!1&&e.push("#define "+t+" "+s)}return e.join(`
`)}function CT(r,e){const t={},s=r.getProgramParameter(e,r.ACTIVE_ATTRIBUTES);for(let a=0;a<s;a++){const l=r.getActiveAttrib(e,a),u=l.name;let f=1;l.type===r.FLOAT_MAT2&&(f=2),l.type===r.FLOAT_MAT3&&(f=3),l.type===r.FLOAT_MAT4&&(f=4),t[u]={type:l.type,location:r.getAttribLocation(e,u),locationSize:f}}return t}function Qo(r){return r!==""}function ng(r,e){const t=e.numSpotLightShadows+e.numSpotLightMaps-e.numSpotLightShadowsWithMaps;return r.replace(/NUM_DIR_LIGHTS/g,e.numDirLights).replace(/NUM_SPOT_LIGHTS/g,e.numSpotLights).replace(/NUM_SPOT_LIGHT_MAPS/g,e.numSpotLightMaps).replace(/NUM_SPOT_LIGHT_COORDS/g,t).replace(/NUM_RECT_AREA_LIGHTS/g,e.numRectAreaLights).replace(/NUM_POINT_LIGHTS/g,e.numPointLights).replace(/NUM_HEMI_LIGHTS/g,e.numHemiLights).replace(/NUM_DIR_LIGHT_SHADOWS/g,e.numDirLightShadows).replace(/NUM_SPOT_LIGHT_SHADOWS_WITH_MAPS/g,e.numSpotLightShadowsWithMaps).replace(/NUM_SPOT_LIGHT_SHADOWS/g,e.numSpotLightShadows).replace(/NUM_POINT_LIGHT_SHADOWS/g,e.numPointLightShadows)}function ig(r,e){return r.replace(/NUM_CLIPPING_PLANES/g,e.numClippingPlanes).replace(/UNION_CLIPPING_PLANES/g,e.numClippingPlanes-e.numClipIntersection)}const RT=/^[ \t]*#include +<([\w\d./]+)>/gm;function fd(r){return r.replace(RT,PT)}const bT=new Map;function PT(r,e){let t=st[e];if(t===void 0){const s=bT.get(e);if(s!==void 0)t=st[s],console.warn('THREE.WebGLRenderer: Shader chunk "%s" has been deprecated. Use "%s" instead.',e,s);else throw new Error("Can not resolve #include <"+e+">")}return fd(t)}const LT=/#pragma unroll_loop_start\s+for\s*\(\s*int\s+i\s*=\s*(\d+)\s*;\s*i\s*<\s*(\d+)\s*;\s*i\s*\+\+\s*\)\s*{([\s\S]+?)}\s+#pragma unroll_loop_end/g;function rg(r){return r.replace(LT,DT)}function DT(r,e,t,s){let a="";for(let l=parseInt(e);l<parseInt(t);l++)a+=s.replace(/\[\s*i\s*\]/g,"[ "+l+" ]").replace(/UNROLLED_LOOP_INDEX/g,l);return a}function sg(r){let e=`precision ${r.precision} float;
	precision ${r.precision} int;
	precision ${r.precision} sampler2D;
	precision ${r.precision} samplerCube;
	precision ${r.precision} sampler3D;
	precision ${r.precision} sampler2DArray;
	precision ${r.precision} sampler2DShadow;
	precision ${r.precision} samplerCubeShadow;
	precision ${r.precision} sampler2DArrayShadow;
	precision ${r.precision} isampler2D;
	precision ${r.precision} isampler3D;
	precision ${r.precision} isamplerCube;
	precision ${r.precision} isampler2DArray;
	precision ${r.precision} usampler2D;
	precision ${r.precision} usampler3D;
	precision ${r.precision} usamplerCube;
	precision ${r.precision} usampler2DArray;
	`;return r.precision==="highp"?e+=`
#define HIGH_PRECISION`:r.precision==="mediump"?e+=`
#define MEDIUM_PRECISION`:r.precision==="lowp"&&(e+=`
#define LOW_PRECISION`),e}function NT(r){let e="SHADOWMAP_TYPE_BASIC";return r.shadowMapType===Dg?e="SHADOWMAP_TYPE_PCF":r.shadowMapType===_x?e="SHADOWMAP_TYPE_PCF_SOFT":r.shadowMapType===Bi&&(e="SHADOWMAP_TYPE_VSM"),e}function IT(r){let e="ENVMAP_TYPE_CUBE";if(r.envMap)switch(r.envMapMode){case to:case no:e="ENVMAP_TYPE_CUBE";break;case tc:e="ENVMAP_TYPE_CUBE_UV";break}return e}function UT(r){let e="ENVMAP_MODE_REFLECTION";return r.envMap&&r.envMapMode===no&&(e="ENVMAP_MODE_REFRACTION"),e}function FT(r){let e="ENVMAP_BLENDING_NONE";if(r.envMap)switch(r.combine){case Ng:e="ENVMAP_BLENDING_MULTIPLY";break;case Fx:e="ENVMAP_BLENDING_MIX";break;case Ox:e="ENVMAP_BLENDING_ADD";break}return e}function OT(r){const e=r.envMapCubeUVHeight;if(e===null)return null;const t=Math.log2(e)-2,s=1/e;return{texelWidth:1/(3*Math.max(Math.pow(2,t),112)),texelHeight:s,maxMip:t}}function kT(r,e,t,s){const a=r.getContext(),l=t.defines;let u=t.vertexShader,f=t.fragmentShader;const h=NT(t),p=IT(t),m=UT(t),_=FT(t),x=OT(t),S=wT(t),T=AT(l),w=a.createProgram();let v,y,P=t.glslVersion?"#version "+t.glslVersion+`
`:"";t.isRawShaderMaterial?(v=["#define SHADER_TYPE "+t.shaderType,"#define SHADER_NAME "+t.shaderName,T].filter(Qo).join(`
`),v.length>0&&(v+=`
`),y=["#define SHADER_TYPE "+t.shaderType,"#define SHADER_NAME "+t.shaderName,T].filter(Qo).join(`
`),y.length>0&&(y+=`
`)):(v=[sg(t),"#define SHADER_TYPE "+t.shaderType,"#define SHADER_NAME "+t.shaderName,T,t.extensionClipCullDistance?"#define USE_CLIP_DISTANCE":"",t.batching?"#define USE_BATCHING":"",t.batchingColor?"#define USE_BATCHING_COLOR":"",t.instancing?"#define USE_INSTANCING":"",t.instancingColor?"#define USE_INSTANCING_COLOR":"",t.instancingMorph?"#define USE_INSTANCING_MORPH":"",t.useFog&&t.fog?"#define USE_FOG":"",t.useFog&&t.fogExp2?"#define FOG_EXP2":"",t.map?"#define USE_MAP":"",t.envMap?"#define USE_ENVMAP":"",t.envMap?"#define "+m:"",t.lightMap?"#define USE_LIGHTMAP":"",t.aoMap?"#define USE_AOMAP":"",t.bumpMap?"#define USE_BUMPMAP":"",t.normalMap?"#define USE_NORMALMAP":"",t.normalMapObjectSpace?"#define USE_NORMALMAP_OBJECTSPACE":"",t.normalMapTangentSpace?"#define USE_NORMALMAP_TANGENTSPACE":"",t.displacementMap?"#define USE_DISPLACEMENTMAP":"",t.emissiveMap?"#define USE_EMISSIVEMAP":"",t.anisotropy?"#define USE_ANISOTROPY":"",t.anisotropyMap?"#define USE_ANISOTROPYMAP":"",t.clearcoatMap?"#define USE_CLEARCOATMAP":"",t.clearcoatRoughnessMap?"#define USE_CLEARCOAT_ROUGHNESSMAP":"",t.clearcoatNormalMap?"#define USE_CLEARCOAT_NORMALMAP":"",t.iridescenceMap?"#define USE_IRIDESCENCEMAP":"",t.iridescenceThicknessMap?"#define USE_IRIDESCENCE_THICKNESSMAP":"",t.specularMap?"#define USE_SPECULARMAP":"",t.specularColorMap?"#define USE_SPECULAR_COLORMAP":"",t.specularIntensityMap?"#define USE_SPECULAR_INTENSITYMAP":"",t.roughnessMap?"#define USE_ROUGHNESSMAP":"",t.metalnessMap?"#define USE_METALNESSMAP":"",t.alphaMap?"#define USE_ALPHAMAP":"",t.alphaHash?"#define USE_ALPHAHASH":"",t.transmission?"#define USE_TRANSMISSION":"",t.transmissionMap?"#define USE_TRANSMISSIONMAP":"",t.thicknessMap?"#define USE_THICKNESSMAP":"",t.sheenColorMap?"#define USE_SHEEN_COLORMAP":"",t.sheenRoughnessMap?"#define USE_SHEEN_ROUGHNESSMAP":"",t.mapUv?"#define MAP_UV "+t.mapUv:"",t.alphaMapUv?"#define ALPHAMAP_UV "+t.alphaMapUv:"",t.lightMapUv?"#define LIGHTMAP_UV "+t.lightMapUv:"",t.aoMapUv?"#define AOMAP_UV "+t.aoMapUv:"",t.emissiveMapUv?"#define EMISSIVEMAP_UV "+t.emissiveMapUv:"",t.bumpMapUv?"#define BUMPMAP_UV "+t.bumpMapUv:"",t.normalMapUv?"#define NORMALMAP_UV "+t.normalMapUv:"",t.displacementMapUv?"#define DISPLACEMENTMAP_UV "+t.displacementMapUv:"",t.metalnessMapUv?"#define METALNESSMAP_UV "+t.metalnessMapUv:"",t.roughnessMapUv?"#define ROUGHNESSMAP_UV "+t.roughnessMapUv:"",t.anisotropyMapUv?"#define ANISOTROPYMAP_UV "+t.anisotropyMapUv:"",t.clearcoatMapUv?"#define CLEARCOATMAP_UV "+t.clearcoatMapUv:"",t.clearcoatNormalMapUv?"#define CLEARCOAT_NORMALMAP_UV "+t.clearcoatNormalMapUv:"",t.clearcoatRoughnessMapUv?"#define CLEARCOAT_ROUGHNESSMAP_UV "+t.clearcoatRoughnessMapUv:"",t.iridescenceMapUv?"#define IRIDESCENCEMAP_UV "+t.iridescenceMapUv:"",t.iridescenceThicknessMapUv?"#define IRIDESCENCE_THICKNESSMAP_UV "+t.iridescenceThicknessMapUv:"",t.sheenColorMapUv?"#define SHEEN_COLORMAP_UV "+t.sheenColorMapUv:"",t.sheenRoughnessMapUv?"#define SHEEN_ROUGHNESSMAP_UV "+t.sheenRoughnessMapUv:"",t.specularMapUv?"#define SPECULARMAP_UV "+t.specularMapUv:"",t.specularColorMapUv?"#define SPECULAR_COLORMAP_UV "+t.specularColorMapUv:"",t.specularIntensityMapUv?"#define SPECULAR_INTENSITYMAP_UV "+t.specularIntensityMapUv:"",t.transmissionMapUv?"#define TRANSMISSIONMAP_UV "+t.transmissionMapUv:"",t.thicknessMapUv?"#define THICKNESSMAP_UV "+t.thicknessMapUv:"",t.vertexTangents&&t.flatShading===!1?"#define USE_TANGENT":"",t.vertexColors?"#define USE_COLOR":"",t.vertexAlphas?"#define USE_COLOR_ALPHA":"",t.vertexUv1s?"#define USE_UV1":"",t.vertexUv2s?"#define USE_UV2":"",t.vertexUv3s?"#define USE_UV3":"",t.pointsUvs?"#define USE_POINTS_UV":"",t.flatShading?"#define FLAT_SHADED":"",t.skinning?"#define USE_SKINNING":"",t.morphTargets?"#define USE_MORPHTARGETS":"",t.morphNormals&&t.flatShading===!1?"#define USE_MORPHNORMALS":"",t.morphColors?"#define USE_MORPHCOLORS":"",t.morphTargetsCount>0?"#define MORPHTARGETS_TEXTURE_STRIDE "+t.morphTextureStride:"",t.morphTargetsCount>0?"#define MORPHTARGETS_COUNT "+t.morphTargetsCount:"",t.doubleSided?"#define DOUBLE_SIDED":"",t.flipSided?"#define FLIP_SIDED":"",t.shadowMapEnabled?"#define USE_SHADOWMAP":"",t.shadowMapEnabled?"#define "+h:"",t.sizeAttenuation?"#define USE_SIZEATTENUATION":"",t.numLightProbes>0?"#define USE_LIGHT_PROBES":"",t.logarithmicDepthBuffer?"#define USE_LOGDEPTHBUF":"",t.reverseDepthBuffer?"#define USE_REVERSEDEPTHBUF":"","uniform mat4 modelMatrix;","uniform mat4 modelViewMatrix;","uniform mat4 projectionMatrix;","uniform mat4 viewMatrix;","uniform mat3 normalMatrix;","uniform vec3 cameraPosition;","uniform bool isOrthographic;","#ifdef USE_INSTANCING","	attribute mat4 instanceMatrix;","#endif","#ifdef USE_INSTANCING_COLOR","	attribute vec3 instanceColor;","#endif","#ifdef USE_INSTANCING_MORPH","	uniform sampler2D morphTexture;","#endif","attribute vec3 position;","attribute vec3 normal;","attribute vec2 uv;","#ifdef USE_UV1","	attribute vec2 uv1;","#endif","#ifdef USE_UV2","	attribute vec2 uv2;","#endif","#ifdef USE_UV3","	attribute vec2 uv3;","#endif","#ifdef USE_TANGENT","	attribute vec4 tangent;","#endif","#if defined( USE_COLOR_ALPHA )","	attribute vec4 color;","#elif defined( USE_COLOR )","	attribute vec3 color;","#endif","#ifdef USE_SKINNING","	attribute vec4 skinIndex;","	attribute vec4 skinWeight;","#endif",`
`].filter(Qo).join(`
`),y=[sg(t),"#define SHADER_TYPE "+t.shaderType,"#define SHADER_NAME "+t.shaderName,T,t.useFog&&t.fog?"#define USE_FOG":"",t.useFog&&t.fogExp2?"#define FOG_EXP2":"",t.alphaToCoverage?"#define ALPHA_TO_COVERAGE":"",t.map?"#define USE_MAP":"",t.matcap?"#define USE_MATCAP":"",t.envMap?"#define USE_ENVMAP":"",t.envMap?"#define "+p:"",t.envMap?"#define "+m:"",t.envMap?"#define "+_:"",x?"#define CUBEUV_TEXEL_WIDTH "+x.texelWidth:"",x?"#define CUBEUV_TEXEL_HEIGHT "+x.texelHeight:"",x?"#define CUBEUV_MAX_MIP "+x.maxMip+".0":"",t.lightMap?"#define USE_LIGHTMAP":"",t.aoMap?"#define USE_AOMAP":"",t.bumpMap?"#define USE_BUMPMAP":"",t.normalMap?"#define USE_NORMALMAP":"",t.normalMapObjectSpace?"#define USE_NORMALMAP_OBJECTSPACE":"",t.normalMapTangentSpace?"#define USE_NORMALMAP_TANGENTSPACE":"",t.emissiveMap?"#define USE_EMISSIVEMAP":"",t.anisotropy?"#define USE_ANISOTROPY":"",t.anisotropyMap?"#define USE_ANISOTROPYMAP":"",t.clearcoat?"#define USE_CLEARCOAT":"",t.clearcoatMap?"#define USE_CLEARCOATMAP":"",t.clearcoatRoughnessMap?"#define USE_CLEARCOAT_ROUGHNESSMAP":"",t.clearcoatNormalMap?"#define USE_CLEARCOAT_NORMALMAP":"",t.dispersion?"#define USE_DISPERSION":"",t.iridescence?"#define USE_IRIDESCENCE":"",t.iridescenceMap?"#define USE_IRIDESCENCEMAP":"",t.iridescenceThicknessMap?"#define USE_IRIDESCENCE_THICKNESSMAP":"",t.specularMap?"#define USE_SPECULARMAP":"",t.specularColorMap?"#define USE_SPECULAR_COLORMAP":"",t.specularIntensityMap?"#define USE_SPECULAR_INTENSITYMAP":"",t.roughnessMap?"#define USE_ROUGHNESSMAP":"",t.metalnessMap?"#define USE_METALNESSMAP":"",t.alphaMap?"#define USE_ALPHAMAP":"",t.alphaTest?"#define USE_ALPHATEST":"",t.alphaHash?"#define USE_ALPHAHASH":"",t.sheen?"#define USE_SHEEN":"",t.sheenColorMap?"#define USE_SHEEN_COLORMAP":"",t.sheenRoughnessMap?"#define USE_SHEEN_ROUGHNESSMAP":"",t.transmission?"#define USE_TRANSMISSION":"",t.transmissionMap?"#define USE_TRANSMISSIONMAP":"",t.thicknessMap?"#define USE_THICKNESSMAP":"",t.vertexTangents&&t.flatShading===!1?"#define USE_TANGENT":"",t.vertexColors||t.instancingColor||t.batchingColor?"#define USE_COLOR":"",t.vertexAlphas?"#define USE_COLOR_ALPHA":"",t.vertexUv1s?"#define USE_UV1":"",t.vertexUv2s?"#define USE_UV2":"",t.vertexUv3s?"#define USE_UV3":"",t.pointsUvs?"#define USE_POINTS_UV":"",t.gradientMap?"#define USE_GRADIENTMAP":"",t.flatShading?"#define FLAT_SHADED":"",t.doubleSided?"#define DOUBLE_SIDED":"",t.flipSided?"#define FLIP_SIDED":"",t.shadowMapEnabled?"#define USE_SHADOWMAP":"",t.shadowMapEnabled?"#define "+h:"",t.premultipliedAlpha?"#define PREMULTIPLIED_ALPHA":"",t.numLightProbes>0?"#define USE_LIGHT_PROBES":"",t.decodeVideoTexture?"#define DECODE_VIDEO_TEXTURE":"",t.logarithmicDepthBuffer?"#define USE_LOGDEPTHBUF":"",t.reverseDepthBuffer?"#define USE_REVERSEDEPTHBUF":"","uniform mat4 viewMatrix;","uniform vec3 cameraPosition;","uniform bool isOrthographic;",t.toneMapping!==wr?"#define TONE_MAPPING":"",t.toneMapping!==wr?st.tonemapping_pars_fragment:"",t.toneMapping!==wr?ET("toneMapping",t.toneMapping):"",t.dithering?"#define DITHERING":"",t.opaque?"#define OPAQUE":"",st.colorspace_pars_fragment,MT("linearToOutputTexel",t.outputColorSpace),TT(),t.useDepthPacking?"#define DEPTH_PACKING "+t.depthPacking:"",`
`].filter(Qo).join(`
`)),u=fd(u),u=ng(u,t),u=ig(u,t),f=fd(f),f=ng(f,t),f=ig(f,t),u=rg(u),f=rg(f),t.isRawShaderMaterial!==!0&&(P=`#version 300 es
`,v=[S,"#define attribute in","#define varying out","#define texture2D texture"].join(`
`)+`
`+v,y=["#define varying in",t.glslVersion===Mm?"":"layout(location = 0) out highp vec4 pc_fragColor;",t.glslVersion===Mm?"":"#define gl_FragColor pc_fragColor","#define gl_FragDepthEXT gl_FragDepth","#define texture2D texture","#define textureCube texture","#define texture2DProj textureProj","#define texture2DLodEXT textureLod","#define texture2DProjLodEXT textureProjLod","#define textureCubeLodEXT textureLod","#define texture2DGradEXT textureGrad","#define texture2DProjGradEXT textureProjGrad","#define textureCubeGradEXT textureGrad"].join(`
`)+`
`+y);const b=P+v+u,D=P+y+f,V=eg(a,a.VERTEX_SHADER,b),O=eg(a,a.FRAGMENT_SHADER,D);a.attachShader(w,V),a.attachShader(w,O),t.index0AttributeName!==void 0?a.bindAttribLocation(w,0,t.index0AttributeName):t.morphTargets===!0&&a.bindAttribLocation(w,0,"position"),a.linkProgram(w);function U(C){if(r.debug.checkShaderErrors){const re=a.getProgramInfoLog(w).trim(),ee=a.getShaderInfoLog(V).trim(),ae=a.getShaderInfoLog(O).trim();let ue=!0,Z=!0;if(a.getProgramParameter(w,a.LINK_STATUS)===!1)if(ue=!1,typeof r.debug.onShaderError=="function")r.debug.onShaderError(a,w,V,O);else{const le=tg(a,V,"vertex"),F=tg(a,O,"fragment");console.error("THREE.WebGLProgram: Shader Error "+a.getError()+" - VALIDATE_STATUS "+a.getProgramParameter(w,a.VALIDATE_STATUS)+`

Material Name: `+C.name+`
Material Type: `+C.type+`

Program Info Log: `+re+`
`+le+`
`+F)}else re!==""?console.warn("THREE.WebGLProgram: Program Info Log:",re):(ee===""||ae==="")&&(Z=!1);Z&&(C.diagnostics={runnable:ue,programLog:re,vertexShader:{log:ee,prefix:v},fragmentShader:{log:ae,prefix:y}})}a.deleteShader(V),a.deleteShader(O),Y=new jl(a,w),ce=CT(a,w)}let Y;this.getUniforms=function(){return Y===void 0&&U(this),Y};let ce;this.getAttributes=function(){return ce===void 0&&U(this),ce};let E=t.rendererExtensionParallelShaderCompile===!1;return this.isReady=function(){return E===!1&&(E=a.getProgramParameter(w,vT)),E},this.destroy=function(){s.releaseStatesOfProgram(this),a.deleteProgram(w),this.program=void 0},this.type=t.shaderType,this.name=t.shaderName,this.id=xT++,this.cacheKey=e,this.usedTimes=1,this.program=w,this.vertexShader=V,this.fragmentShader=O,this}let BT=0;class zT{constructor(){this.shaderCache=new Map,this.materialCache=new Map}update(e){const t=e.vertexShader,s=e.fragmentShader,a=this._getShaderStage(t),l=this._getShaderStage(s),u=this._getShaderCacheForMaterial(e);return u.has(a)===!1&&(u.add(a),a.usedTimes++),u.has(l)===!1&&(u.add(l),l.usedTimes++),this}remove(e){const t=this.materialCache.get(e);for(const s of t)s.usedTimes--,s.usedTimes===0&&this.shaderCache.delete(s.code);return this.materialCache.delete(e),this}getVertexShaderID(e){return this._getShaderStage(e.vertexShader).id}getFragmentShaderID(e){return this._getShaderStage(e.fragmentShader).id}dispose(){this.shaderCache.clear(),this.materialCache.clear()}_getShaderCacheForMaterial(e){const t=this.materialCache;let s=t.get(e);return s===void 0&&(s=new Set,t.set(e,s)),s}_getShaderStage(e){const t=this.shaderCache;let s=t.get(e);return s===void 0&&(s=new HT(e),t.set(e,s)),s}}class HT{constructor(e){this.id=BT++,this.code=e,this.usedTimes=0}}function VT(r,e,t,s,a,l,u){const f=new Zg,h=new zT,p=new Set,m=[],_=a.logarithmicDepthBuffer,x=a.reverseDepthBuffer,S=a.vertexTextures;let T=a.precision;const w={MeshDepthMaterial:"depth",MeshDistanceMaterial:"distanceRGBA",MeshNormalMaterial:"normal",MeshBasicMaterial:"basic",MeshLambertMaterial:"lambert",MeshPhongMaterial:"phong",MeshToonMaterial:"toon",MeshStandardMaterial:"physical",MeshPhysicalMaterial:"physical",MeshMatcapMaterial:"matcap",LineBasicMaterial:"basic",LineDashedMaterial:"dashed",PointsMaterial:"points",ShadowMaterial:"shadow",SpriteMaterial:"sprite"};function v(E){return p.add(E),E===0?"uv":`uv${E}`}function y(E,C,re,ee,ae){const ue=ee.fog,Z=ae.geometry,le=E.isMeshStandardMaterial?ee.environment:null,F=(E.isMeshStandardMaterial?t:e).get(E.envMap||le),se=F&&F.mapping===tc?F.image.height:null,L=w[E.type];E.precision!==null&&(T=a.getMaxPrecision(E.precision),T!==E.precision&&console.warn("THREE.WebGLProgram.getParameters:",E.precision,"not supported, using",T,"instead."));const X=Z.morphAttributes.position||Z.morphAttributes.normal||Z.morphAttributes.color,ve=X!==void 0?X.length:0;let Ne=0;Z.morphAttributes.position!==void 0&&(Ne=1),Z.morphAttributes.normal!==void 0&&(Ne=2),Z.morphAttributes.color!==void 0&&(Ne=3);let J,fe,Se,Me;if(L){const en=vi[L];J=en.vertexShader,fe=en.fragmentShader}else J=E.vertexShader,fe=E.fragmentShader,h.update(E),Se=h.getVertexShaderID(E),Me=h.getFragmentShaderID(E);const Pe=r.getRenderTarget(),Ge=ae.isInstancedMesh===!0,dt=ae.isBatchedMesh===!0,gt=!!E.map,ct=!!E.matcap,z=!!F,an=!!E.aoMap,at=!!E.lightMap,ht=!!E.bumpMap,Ze=!!E.normalMap,At=!!E.displacementMap,Je=!!E.emissiveMap,N=!!E.metalnessMap,A=!!E.roughnessMap,K=E.anisotropy>0,pe=E.clearcoat>0,xe=E.dispersion>0,de=E.iridescence>0,Ye=E.sheen>0,Ce=E.transmission>0,Fe=K&&!!E.anisotropyMap,pt=pe&&!!E.clearcoatMap,Ee=pe&&!!E.clearcoatNormalMap,Oe=pe&&!!E.clearcoatRoughnessMap,tt=de&&!!E.iridescenceMap,et=de&&!!E.iridescenceThicknessMap,Be=Ye&&!!E.sheenColorMap,ut=Ye&&!!E.sheenRoughnessMap,it=!!E.specularMap,Et=!!E.specularColorMap,G=!!E.specularIntensityMap,Le=Ce&&!!E.transmissionMap,oe=Ce&&!!E.thicknessMap,me=!!E.gradientMap,Re=!!E.alphaMap,Ie=E.alphaTest>0,ft=!!E.alphaHash,kt=!!E.extensions;let ln=wr;E.toneMapped&&(Pe===null||Pe.isXRRenderTarget===!0)&&(ln=r.toneMapping);const mt={shaderID:L,shaderType:E.type,shaderName:E.name,vertexShader:J,fragmentShader:fe,defines:E.defines,customVertexShaderID:Se,customFragmentShaderID:Me,isRawShaderMaterial:E.isRawShaderMaterial===!0,glslVersion:E.glslVersion,precision:T,batching:dt,batchingColor:dt&&ae._colorsTexture!==null,instancing:Ge,instancingColor:Ge&&ae.instanceColor!==null,instancingMorph:Ge&&ae.morphTexture!==null,supportsVertexTextures:S,outputColorSpace:Pe===null?r.outputColorSpace:Pe.isXRRenderTarget===!0?Pe.texture.colorSpace:br,alphaToCoverage:!!E.alphaToCoverage,map:gt,matcap:ct,envMap:z,envMapMode:z&&F.mapping,envMapCubeUVHeight:se,aoMap:an,lightMap:at,bumpMap:ht,normalMap:Ze,displacementMap:S&&At,emissiveMap:Je,normalMapObjectSpace:Ze&&E.normalMapType===qx,normalMapTangentSpace:Ze&&E.normalMapType===jg,metalnessMap:N,roughnessMap:A,anisotropy:K,anisotropyMap:Fe,clearcoat:pe,clearcoatMap:pt,clearcoatNormalMap:Ee,clearcoatRoughnessMap:Oe,dispersion:xe,iridescence:de,iridescenceMap:tt,iridescenceThicknessMap:et,sheen:Ye,sheenColorMap:Be,sheenRoughnessMap:ut,specularMap:it,specularColorMap:Et,specularIntensityMap:G,transmission:Ce,transmissionMap:Le,thicknessMap:oe,gradientMap:me,opaque:E.transparent===!1&&E.blending===Zs&&E.alphaToCoverage===!1,alphaMap:Re,alphaTest:Ie,alphaHash:ft,combine:E.combine,mapUv:gt&&v(E.map.channel),aoMapUv:an&&v(E.aoMap.channel),lightMapUv:at&&v(E.lightMap.channel),bumpMapUv:ht&&v(E.bumpMap.channel),normalMapUv:Ze&&v(E.normalMap.channel),displacementMapUv:At&&v(E.displacementMap.channel),emissiveMapUv:Je&&v(E.emissiveMap.channel),metalnessMapUv:N&&v(E.metalnessMap.channel),roughnessMapUv:A&&v(E.roughnessMap.channel),anisotropyMapUv:Fe&&v(E.anisotropyMap.channel),clearcoatMapUv:pt&&v(E.clearcoatMap.channel),clearcoatNormalMapUv:Ee&&v(E.clearcoatNormalMap.channel),clearcoatRoughnessMapUv:Oe&&v(E.clearcoatRoughnessMap.channel),iridescenceMapUv:tt&&v(E.iridescenceMap.channel),iridescenceThicknessMapUv:et&&v(E.iridescenceThicknessMap.channel),sheenColorMapUv:Be&&v(E.sheenColorMap.channel),sheenRoughnessMapUv:ut&&v(E.sheenRoughnessMap.channel),specularMapUv:it&&v(E.specularMap.channel),specularColorMapUv:Et&&v(E.specularColorMap.channel),specularIntensityMapUv:G&&v(E.specularIntensityMap.channel),transmissionMapUv:Le&&v(E.transmissionMap.channel),thicknessMapUv:oe&&v(E.thicknessMap.channel),alphaMapUv:Re&&v(E.alphaMap.channel),vertexTangents:!!Z.attributes.tangent&&(Ze||K),vertexColors:E.vertexColors,vertexAlphas:E.vertexColors===!0&&!!Z.attributes.color&&Z.attributes.color.itemSize===4,pointsUvs:ae.isPoints===!0&&!!Z.attributes.uv&&(gt||Re),fog:!!ue,useFog:E.fog===!0,fogExp2:!!ue&&ue.isFogExp2,flatShading:E.flatShading===!0,sizeAttenuation:E.sizeAttenuation===!0,logarithmicDepthBuffer:_,reverseDepthBuffer:x,skinning:ae.isSkinnedMesh===!0,morphTargets:Z.morphAttributes.position!==void 0,morphNormals:Z.morphAttributes.normal!==void 0,morphColors:Z.morphAttributes.color!==void 0,morphTargetsCount:ve,morphTextureStride:Ne,numDirLights:C.directional.length,numPointLights:C.point.length,numSpotLights:C.spot.length,numSpotLightMaps:C.spotLightMap.length,numRectAreaLights:C.rectArea.length,numHemiLights:C.hemi.length,numDirLightShadows:C.directionalShadowMap.length,numPointLightShadows:C.pointShadowMap.length,numSpotLightShadows:C.spotShadowMap.length,numSpotLightShadowsWithMaps:C.numSpotLightShadowsWithMaps,numLightProbes:C.numLightProbes,numClippingPlanes:u.numPlanes,numClipIntersection:u.numIntersection,dithering:E.dithering,shadowMapEnabled:r.shadowMap.enabled&&re.length>0,shadowMapType:r.shadowMap.type,toneMapping:ln,decodeVideoTexture:gt&&E.map.isVideoTexture===!0&&Tt.getTransfer(E.map.colorSpace)===Ut,premultipliedAlpha:E.premultipliedAlpha,doubleSided:E.side===zi,flipSided:E.side===Ln,useDepthPacking:E.depthPacking>=0,depthPacking:E.depthPacking||0,index0AttributeName:E.index0AttributeName,extensionClipCullDistance:kt&&E.extensions.clipCullDistance===!0&&s.has("WEBGL_clip_cull_distance"),extensionMultiDraw:(kt&&E.extensions.multiDraw===!0||dt)&&s.has("WEBGL_multi_draw"),rendererExtensionParallelShaderCompile:s.has("KHR_parallel_shader_compile"),customProgramCacheKey:E.customProgramCacheKey()};return mt.vertexUv1s=p.has(1),mt.vertexUv2s=p.has(2),mt.vertexUv3s=p.has(3),p.clear(),mt}function P(E){const C=[];if(E.shaderID?C.push(E.shaderID):(C.push(E.customVertexShaderID),C.push(E.customFragmentShaderID)),E.defines!==void 0)for(const re in E.defines)C.push(re),C.push(E.defines[re]);return E.isRawShaderMaterial===!1&&(b(C,E),D(C,E),C.push(r.outputColorSpace)),C.push(E.customProgramCacheKey),C.join()}function b(E,C){E.push(C.precision),E.push(C.outputColorSpace),E.push(C.envMapMode),E.push(C.envMapCubeUVHeight),E.push(C.mapUv),E.push(C.alphaMapUv),E.push(C.lightMapUv),E.push(C.aoMapUv),E.push(C.bumpMapUv),E.push(C.normalMapUv),E.push(C.displacementMapUv),E.push(C.emissiveMapUv),E.push(C.metalnessMapUv),E.push(C.roughnessMapUv),E.push(C.anisotropyMapUv),E.push(C.clearcoatMapUv),E.push(C.clearcoatNormalMapUv),E.push(C.clearcoatRoughnessMapUv),E.push(C.iridescenceMapUv),E.push(C.iridescenceThicknessMapUv),E.push(C.sheenColorMapUv),E.push(C.sheenRoughnessMapUv),E.push(C.specularMapUv),E.push(C.specularColorMapUv),E.push(C.specularIntensityMapUv),E.push(C.transmissionMapUv),E.push(C.thicknessMapUv),E.push(C.combine),E.push(C.fogExp2),E.push(C.sizeAttenuation),E.push(C.morphTargetsCount),E.push(C.morphAttributeCount),E.push(C.numDirLights),E.push(C.numPointLights),E.push(C.numSpotLights),E.push(C.numSpotLightMaps),E.push(C.numHemiLights),E.push(C.numRectAreaLights),E.push(C.numDirLightShadows),E.push(C.numPointLightShadows),E.push(C.numSpotLightShadows),E.push(C.numSpotLightShadowsWithMaps),E.push(C.numLightProbes),E.push(C.shadowMapType),E.push(C.toneMapping),E.push(C.numClippingPlanes),E.push(C.numClipIntersection),E.push(C.depthPacking)}function D(E,C){f.disableAll(),C.supportsVertexTextures&&f.enable(0),C.instancing&&f.enable(1),C.instancingColor&&f.enable(2),C.instancingMorph&&f.enable(3),C.matcap&&f.enable(4),C.envMap&&f.enable(5),C.normalMapObjectSpace&&f.enable(6),C.normalMapTangentSpace&&f.enable(7),C.clearcoat&&f.enable(8),C.iridescence&&f.enable(9),C.alphaTest&&f.enable(10),C.vertexColors&&f.enable(11),C.vertexAlphas&&f.enable(12),C.vertexUv1s&&f.enable(13),C.vertexUv2s&&f.enable(14),C.vertexUv3s&&f.enable(15),C.vertexTangents&&f.enable(16),C.anisotropy&&f.enable(17),C.alphaHash&&f.enable(18),C.batching&&f.enable(19),C.dispersion&&f.enable(20),C.batchingColor&&f.enable(21),E.push(f.mask),f.disableAll(),C.fog&&f.enable(0),C.useFog&&f.enable(1),C.flatShading&&f.enable(2),C.logarithmicDepthBuffer&&f.enable(3),C.reverseDepthBuffer&&f.enable(4),C.skinning&&f.enable(5),C.morphTargets&&f.enable(6),C.morphNormals&&f.enable(7),C.morphColors&&f.enable(8),C.premultipliedAlpha&&f.enable(9),C.shadowMapEnabled&&f.enable(10),C.doubleSided&&f.enable(11),C.flipSided&&f.enable(12),C.useDepthPacking&&f.enable(13),C.dithering&&f.enable(14),C.transmission&&f.enable(15),C.sheen&&f.enable(16),C.opaque&&f.enable(17),C.pointsUvs&&f.enable(18),C.decodeVideoTexture&&f.enable(19),C.alphaToCoverage&&f.enable(20),E.push(f.mask)}function V(E){const C=w[E.type];let re;if(C){const ee=vi[C];re=wy.clone(ee.uniforms)}else re=E.uniforms;return re}function O(E,C){let re;for(let ee=0,ae=m.length;ee<ae;ee++){const ue=m[ee];if(ue.cacheKey===C){re=ue,++re.usedTimes;break}}return re===void 0&&(re=new kT(r,C,E,l),m.push(re)),re}function U(E){if(--E.usedTimes===0){const C=m.indexOf(E);m[C]=m[m.length-1],m.pop(),E.destroy()}}function Y(E){h.remove(E)}function ce(){h.dispose()}return{getParameters:y,getProgramCacheKey:P,getUniforms:V,acquireProgram:O,releaseProgram:U,releaseShaderCache:Y,programs:m,dispose:ce}}function GT(){let r=new WeakMap;function e(u){return r.has(u)}function t(u){let f=r.get(u);return f===void 0&&(f={},r.set(u,f)),f}function s(u){r.delete(u)}function a(u,f,h){r.get(u)[f]=h}function l(){r=new WeakMap}return{has:e,get:t,remove:s,update:a,dispose:l}}function WT(r,e){return r.groupOrder!==e.groupOrder?r.groupOrder-e.groupOrder:r.renderOrder!==e.renderOrder?r.renderOrder-e.renderOrder:r.material.id!==e.material.id?r.material.id-e.material.id:r.z!==e.z?r.z-e.z:r.id-e.id}function og(r,e){return r.groupOrder!==e.groupOrder?r.groupOrder-e.groupOrder:r.renderOrder!==e.renderOrder?r.renderOrder-e.renderOrder:r.z!==e.z?e.z-r.z:r.id-e.id}function ag(){const r=[];let e=0;const t=[],s=[],a=[];function l(){e=0,t.length=0,s.length=0,a.length=0}function u(_,x,S,T,w,v){let y=r[e];return y===void 0?(y={id:_.id,object:_,geometry:x,material:S,groupOrder:T,renderOrder:_.renderOrder,z:w,group:v},r[e]=y):(y.id=_.id,y.object=_,y.geometry=x,y.material=S,y.groupOrder=T,y.renderOrder=_.renderOrder,y.z=w,y.group=v),e++,y}function f(_,x,S,T,w,v){const y=u(_,x,S,T,w,v);S.transmission>0?s.push(y):S.transparent===!0?a.push(y):t.push(y)}function h(_,x,S,T,w,v){const y=u(_,x,S,T,w,v);S.transmission>0?s.unshift(y):S.transparent===!0?a.unshift(y):t.unshift(y)}function p(_,x){t.length>1&&t.sort(_||WT),s.length>1&&s.sort(x||og),a.length>1&&a.sort(x||og)}function m(){for(let _=e,x=r.length;_<x;_++){const S=r[_];if(S.id===null)break;S.id=null,S.object=null,S.geometry=null,S.material=null,S.group=null}}return{opaque:t,transmissive:s,transparent:a,init:l,push:f,unshift:h,finish:m,sort:p}}function jT(){let r=new WeakMap;function e(s,a){const l=r.get(s);let u;return l===void 0?(u=new ag,r.set(s,[u])):a>=l.length?(u=new ag,l.push(u)):u=l[a],u}function t(){r=new WeakMap}return{get:e,dispose:t}}function XT(){const r={};return{get:function(e){if(r[e.id]!==void 0)return r[e.id];let t;switch(e.type){case"DirectionalLight":t={direction:new Q,color:new vt};break;case"SpotLight":t={position:new Q,direction:new Q,color:new vt,distance:0,coneCos:0,penumbraCos:0,decay:0};break;case"PointLight":t={position:new Q,color:new vt,distance:0,decay:0};break;case"HemisphereLight":t={direction:new Q,skyColor:new vt,groundColor:new vt};break;case"RectAreaLight":t={color:new vt,position:new Q,halfWidth:new Q,halfHeight:new Q};break}return r[e.id]=t,t}}}function YT(){const r={};return{get:function(e){if(r[e.id]!==void 0)return r[e.id];let t;switch(e.type){case"DirectionalLight":t={shadowIntensity:1,shadowBias:0,shadowNormalBias:0,shadowRadius:1,shadowMapSize:new rt};break;case"SpotLight":t={shadowIntensity:1,shadowBias:0,shadowNormalBias:0,shadowRadius:1,shadowMapSize:new rt};break;case"PointLight":t={shadowIntensity:1,shadowBias:0,shadowNormalBias:0,shadowRadius:1,shadowMapSize:new rt,shadowCameraNear:1,shadowCameraFar:1e3};break}return r[e.id]=t,t}}}let qT=0;function $T(r,e){return(e.castShadow?2:0)-(r.castShadow?2:0)+(e.map?1:0)-(r.map?1:0)}function KT(r){const e=new XT,t=YT(),s={version:0,hash:{directionalLength:-1,pointLength:-1,spotLength:-1,rectAreaLength:-1,hemiLength:-1,numDirectionalShadows:-1,numPointShadows:-1,numSpotShadows:-1,numSpotMaps:-1,numLightProbes:-1},ambient:[0,0,0],probe:[],directional:[],directionalShadow:[],directionalShadowMap:[],directionalShadowMatrix:[],spot:[],spotLightMap:[],spotShadow:[],spotShadowMap:[],spotLightMatrix:[],rectArea:[],rectAreaLTC1:null,rectAreaLTC2:null,point:[],pointShadow:[],pointShadowMap:[],pointShadowMatrix:[],hemi:[],numSpotLightShadowsWithMaps:0,numLightProbes:0};for(let p=0;p<9;p++)s.probe.push(new Q);const a=new Q,l=new Gt,u=new Gt;function f(p){let m=0,_=0,x=0;for(let ce=0;ce<9;ce++)s.probe[ce].set(0,0,0);let S=0,T=0,w=0,v=0,y=0,P=0,b=0,D=0,V=0,O=0,U=0;p.sort($T);for(let ce=0,E=p.length;ce<E;ce++){const C=p[ce],re=C.color,ee=C.intensity,ae=C.distance,ue=C.shadow&&C.shadow.map?C.shadow.map.texture:null;if(C.isAmbientLight)m+=re.r*ee,_+=re.g*ee,x+=re.b*ee;else if(C.isLightProbe){for(let Z=0;Z<9;Z++)s.probe[Z].addScaledVector(C.sh.coefficients[Z],ee);U++}else if(C.isDirectionalLight){const Z=e.get(C);if(Z.color.copy(C.color).multiplyScalar(C.intensity),C.castShadow){const le=C.shadow,F=t.get(C);F.shadowIntensity=le.intensity,F.shadowBias=le.bias,F.shadowNormalBias=le.normalBias,F.shadowRadius=le.radius,F.shadowMapSize=le.mapSize,s.directionalShadow[S]=F,s.directionalShadowMap[S]=ue,s.directionalShadowMatrix[S]=C.shadow.matrix,P++}s.directional[S]=Z,S++}else if(C.isSpotLight){const Z=e.get(C);Z.position.setFromMatrixPosition(C.matrixWorld),Z.color.copy(re).multiplyScalar(ee),Z.distance=ae,Z.coneCos=Math.cos(C.angle),Z.penumbraCos=Math.cos(C.angle*(1-C.penumbra)),Z.decay=C.decay,s.spot[w]=Z;const le=C.shadow;if(C.map&&(s.spotLightMap[V]=C.map,V++,le.updateMatrices(C),C.castShadow&&O++),s.spotLightMatrix[w]=le.matrix,C.castShadow){const F=t.get(C);F.shadowIntensity=le.intensity,F.shadowBias=le.bias,F.shadowNormalBias=le.normalBias,F.shadowRadius=le.radius,F.shadowMapSize=le.mapSize,s.spotShadow[w]=F,s.spotShadowMap[w]=ue,D++}w++}else if(C.isRectAreaLight){const Z=e.get(C);Z.color.copy(re).multiplyScalar(ee),Z.halfWidth.set(C.width*.5,0,0),Z.halfHeight.set(0,C.height*.5,0),s.rectArea[v]=Z,v++}else if(C.isPointLight){const Z=e.get(C);if(Z.color.copy(C.color).multiplyScalar(C.intensity),Z.distance=C.distance,Z.decay=C.decay,C.castShadow){const le=C.shadow,F=t.get(C);F.shadowIntensity=le.intensity,F.shadowBias=le.bias,F.shadowNormalBias=le.normalBias,F.shadowRadius=le.radius,F.shadowMapSize=le.mapSize,F.shadowCameraNear=le.camera.near,F.shadowCameraFar=le.camera.far,s.pointShadow[T]=F,s.pointShadowMap[T]=ue,s.pointShadowMatrix[T]=C.shadow.matrix,b++}s.point[T]=Z,T++}else if(C.isHemisphereLight){const Z=e.get(C);Z.skyColor.copy(C.color).multiplyScalar(ee),Z.groundColor.copy(C.groundColor).multiplyScalar(ee),s.hemi[y]=Z,y++}}v>0&&(r.has("OES_texture_float_linear")===!0?(s.rectAreaLTC1=be.LTC_FLOAT_1,s.rectAreaLTC2=be.LTC_FLOAT_2):(s.rectAreaLTC1=be.LTC_HALF_1,s.rectAreaLTC2=be.LTC_HALF_2)),s.ambient[0]=m,s.ambient[1]=_,s.ambient[2]=x;const Y=s.hash;(Y.directionalLength!==S||Y.pointLength!==T||Y.spotLength!==w||Y.rectAreaLength!==v||Y.hemiLength!==y||Y.numDirectionalShadows!==P||Y.numPointShadows!==b||Y.numSpotShadows!==D||Y.numSpotMaps!==V||Y.numLightProbes!==U)&&(s.directional.length=S,s.spot.length=w,s.rectArea.length=v,s.point.length=T,s.hemi.length=y,s.directionalShadow.length=P,s.directionalShadowMap.length=P,s.pointShadow.length=b,s.pointShadowMap.length=b,s.spotShadow.length=D,s.spotShadowMap.length=D,s.directionalShadowMatrix.length=P,s.pointShadowMatrix.length=b,s.spotLightMatrix.length=D+V-O,s.spotLightMap.length=V,s.numSpotLightShadowsWithMaps=O,s.numLightProbes=U,Y.directionalLength=S,Y.pointLength=T,Y.spotLength=w,Y.rectAreaLength=v,Y.hemiLength=y,Y.numDirectionalShadows=P,Y.numPointShadows=b,Y.numSpotShadows=D,Y.numSpotMaps=V,Y.numLightProbes=U,s.version=qT++)}function h(p,m){let _=0,x=0,S=0,T=0,w=0;const v=m.matrixWorldInverse;for(let y=0,P=p.length;y<P;y++){const b=p[y];if(b.isDirectionalLight){const D=s.directional[_];D.direction.setFromMatrixPosition(b.matrixWorld),a.setFromMatrixPosition(b.target.matrixWorld),D.direction.sub(a),D.direction.transformDirection(v),_++}else if(b.isSpotLight){const D=s.spot[S];D.position.setFromMatrixPosition(b.matrixWorld),D.position.applyMatrix4(v),D.direction.setFromMatrixPosition(b.matrixWorld),a.setFromMatrixPosition(b.target.matrixWorld),D.direction.sub(a),D.direction.transformDirection(v),S++}else if(b.isRectAreaLight){const D=s.rectArea[T];D.position.setFromMatrixPosition(b.matrixWorld),D.position.applyMatrix4(v),u.identity(),l.copy(b.matrixWorld),l.premultiply(v),u.extractRotation(l),D.halfWidth.set(b.width*.5,0,0),D.halfHeight.set(0,b.height*.5,0),D.halfWidth.applyMatrix4(u),D.halfHeight.applyMatrix4(u),T++}else if(b.isPointLight){const D=s.point[x];D.position.setFromMatrixPosition(b.matrixWorld),D.position.applyMatrix4(v),x++}else if(b.isHemisphereLight){const D=s.hemi[w];D.direction.setFromMatrixPosition(b.matrixWorld),D.direction.transformDirection(v),w++}}}return{setup:f,setupView:h,state:s}}function lg(r){const e=new KT(r),t=[],s=[];function a(m){p.camera=m,t.length=0,s.length=0}function l(m){t.push(m)}function u(m){s.push(m)}function f(){e.setup(t)}function h(m){e.setupView(t,m)}const p={lightsArray:t,shadowsArray:s,camera:null,lights:e,transmissionRenderTarget:{}};return{init:a,state:p,setupLights:f,setupLightsView:h,pushLight:l,pushShadow:u}}function ZT(r){let e=new WeakMap;function t(a,l=0){const u=e.get(a);let f;return u===void 0?(f=new lg(r),e.set(a,[f])):l>=u.length?(f=new lg(r),u.push(f)):f=u[l],f}function s(){e=new WeakMap}return{get:t,dispose:s}}class QT extends aa{constructor(e){super(),this.isMeshDepthMaterial=!0,this.type="MeshDepthMaterial",this.depthPacking=Xx,this.map=null,this.alphaMap=null,this.displacementMap=null,this.displacementScale=1,this.displacementBias=0,this.wireframe=!1,this.wireframeLinewidth=1,this.setValues(e)}copy(e){return super.copy(e),this.depthPacking=e.depthPacking,this.map=e.map,this.alphaMap=e.alphaMap,this.displacementMap=e.displacementMap,this.displacementScale=e.displacementScale,this.displacementBias=e.displacementBias,this.wireframe=e.wireframe,this.wireframeLinewidth=e.wireframeLinewidth,this}}class JT extends aa{constructor(e){super(),this.isMeshDistanceMaterial=!0,this.type="MeshDistanceMaterial",this.map=null,this.alphaMap=null,this.displacementMap=null,this.displacementScale=1,this.displacementBias=0,this.setValues(e)}copy(e){return super.copy(e),this.map=e.map,this.alphaMap=e.alphaMap,this.displacementMap=e.displacementMap,this.displacementScale=e.displacementScale,this.displacementBias=e.displacementBias,this}}const ew=`void main() {
	gl_Position = vec4( position, 1.0 );
}`,tw=`uniform sampler2D shadow_pass;
uniform vec2 resolution;
uniform float radius;
#include <packing>
void main() {
	const float samples = float( VSM_SAMPLES );
	float mean = 0.0;
	float squared_mean = 0.0;
	float uvStride = samples <= 1.0 ? 0.0 : 2.0 / ( samples - 1.0 );
	float uvStart = samples <= 1.0 ? 0.0 : - 1.0;
	for ( float i = 0.0; i < samples; i ++ ) {
		float uvOffset = uvStart + i * uvStride;
		#ifdef HORIZONTAL_PASS
			vec2 distribution = unpackRGBATo2Half( texture2D( shadow_pass, ( gl_FragCoord.xy + vec2( uvOffset, 0.0 ) * radius ) / resolution ) );
			mean += distribution.x;
			squared_mean += distribution.y * distribution.y + distribution.x * distribution.x;
		#else
			float depth = unpackRGBAToDepth( texture2D( shadow_pass, ( gl_FragCoord.xy + vec2( 0.0, uvOffset ) * radius ) / resolution ) );
			mean += depth;
			squared_mean += depth * depth;
		#endif
	}
	mean = mean / samples;
	squared_mean = squared_mean / samples;
	float std_dev = sqrt( squared_mean - mean * mean );
	gl_FragColor = pack2HalfToRGBA( vec2( mean, std_dev ) );
}`;function nw(r,e,t){let s=new wd;const a=new rt,l=new rt,u=new Vt,f=new QT({depthPacking:Yx}),h=new JT,p={},m=t.maxTextureSize,_={[Ar]:Ln,[Ln]:Ar,[zi]:zi},x=new Cr({defines:{VSM_SAMPLES:8},uniforms:{shadow_pass:{value:null},resolution:{value:new rt},radius:{value:4}},vertexShader:ew,fragmentShader:tw}),S=x.clone();S.defines.HORIZONTAL_PASS=1;const T=new ji;T.setAttribute("position",new Vn(new Float32Array([-1,-1,.5,3,-1,.5,-1,3,.5]),3));const w=new xi(T,x),v=this;this.enabled=!1,this.autoUpdate=!0,this.needsUpdate=!1,this.type=Dg;let y=this.type;this.render=function(O,U,Y){if(v.enabled===!1||v.autoUpdate===!1&&v.needsUpdate===!1||O.length===0)return;const ce=r.getRenderTarget(),E=r.getActiveCubeFace(),C=r.getActiveMipmapLevel(),re=r.state;re.setBlending(Tr),re.buffers.color.setClear(1,1,1,1),re.buffers.depth.setTest(!0),re.setScissorTest(!1);const ee=y!==Bi&&this.type===Bi,ae=y===Bi&&this.type!==Bi;for(let ue=0,Z=O.length;ue<Z;ue++){const le=O[ue],F=le.shadow;if(F===void 0){console.warn("THREE.WebGLShadowMap:",le,"has no shadow.");continue}if(F.autoUpdate===!1&&F.needsUpdate===!1)continue;a.copy(F.mapSize);const se=F.getFrameExtents();if(a.multiply(se),l.copy(F.mapSize),(a.x>m||a.y>m)&&(a.x>m&&(l.x=Math.floor(m/se.x),a.x=l.x*se.x,F.mapSize.x=l.x),a.y>m&&(l.y=Math.floor(m/se.y),a.y=l.y*se.y,F.mapSize.y=l.y)),F.map===null||ee===!0||ae===!0){const X=this.type!==Bi?{minFilter:Qn,magFilter:Qn}:{};F.map!==null&&F.map.dispose(),F.map=new is(a.x,a.y,X),F.map.texture.name=le.name+".shadowMap",F.camera.updateProjectionMatrix()}r.setRenderTarget(F.map),r.clear();const L=F.getViewportCount();for(let X=0;X<L;X++){const ve=F.getViewport(X);u.set(l.x*ve.x,l.y*ve.y,l.x*ve.z,l.y*ve.w),re.viewport(u),F.updateMatrices(le,X),s=F.getFrustum(),D(U,Y,F.camera,le,this.type)}F.isPointLightShadow!==!0&&this.type===Bi&&P(F,Y),F.needsUpdate=!1}y=this.type,v.needsUpdate=!1,r.setRenderTarget(ce,E,C)};function P(O,U){const Y=e.update(w);x.defines.VSM_SAMPLES!==O.blurSamples&&(x.defines.VSM_SAMPLES=O.blurSamples,S.defines.VSM_SAMPLES=O.blurSamples,x.needsUpdate=!0,S.needsUpdate=!0),O.mapPass===null&&(O.mapPass=new is(a.x,a.y)),x.uniforms.shadow_pass.value=O.map.texture,x.uniforms.resolution.value=O.mapSize,x.uniforms.radius.value=O.radius,r.setRenderTarget(O.mapPass),r.clear(),r.renderBufferDirect(U,null,Y,x,w,null),S.uniforms.shadow_pass.value=O.mapPass.texture,S.uniforms.resolution.value=O.mapSize,S.uniforms.radius.value=O.radius,r.setRenderTarget(O.map),r.clear(),r.renderBufferDirect(U,null,Y,S,w,null)}function b(O,U,Y,ce){let E=null;const C=Y.isPointLight===!0?O.customDistanceMaterial:O.customDepthMaterial;if(C!==void 0)E=C;else if(E=Y.isPointLight===!0?h:f,r.localClippingEnabled&&U.clipShadows===!0&&Array.isArray(U.clippingPlanes)&&U.clippingPlanes.length!==0||U.displacementMap&&U.displacementScale!==0||U.alphaMap&&U.alphaTest>0||U.map&&U.alphaTest>0){const re=E.uuid,ee=U.uuid;let ae=p[re];ae===void 0&&(ae={},p[re]=ae);let ue=ae[ee];ue===void 0&&(ue=E.clone(),ae[ee]=ue,U.addEventListener("dispose",V)),E=ue}if(E.visible=U.visible,E.wireframe=U.wireframe,ce===Bi?E.side=U.shadowSide!==null?U.shadowSide:U.side:E.side=U.shadowSide!==null?U.shadowSide:_[U.side],E.alphaMap=U.alphaMap,E.alphaTest=U.alphaTest,E.map=U.map,E.clipShadows=U.clipShadows,E.clippingPlanes=U.clippingPlanes,E.clipIntersection=U.clipIntersection,E.displacementMap=U.displacementMap,E.displacementScale=U.displacementScale,E.displacementBias=U.displacementBias,E.wireframeLinewidth=U.wireframeLinewidth,E.linewidth=U.linewidth,Y.isPointLight===!0&&E.isMeshDistanceMaterial===!0){const re=r.properties.get(E);re.light=Y}return E}function D(O,U,Y,ce,E){if(O.visible===!1)return;if(O.layers.test(U.layers)&&(O.isMesh||O.isLine||O.isPoints)&&(O.castShadow||O.receiveShadow&&E===Bi)&&(!O.frustumCulled||s.intersectsObject(O))){O.modelViewMatrix.multiplyMatrices(Y.matrixWorldInverse,O.matrixWorld);const ee=e.update(O),ae=O.material;if(Array.isArray(ae)){const ue=ee.groups;for(let Z=0,le=ue.length;Z<le;Z++){const F=ue[Z],se=ae[F.materialIndex];if(se&&se.visible){const L=b(O,se,ce,E);O.onBeforeShadow(r,O,U,Y,ee,L,F),r.renderBufferDirect(Y,null,ee,L,O,F),O.onAfterShadow(r,O,U,Y,ee,L,F)}}}else if(ae.visible){const ue=b(O,ae,ce,E);O.onBeforeShadow(r,O,U,Y,ee,ue,null),r.renderBufferDirect(Y,null,ee,ue,O,null),O.onAfterShadow(r,O,U,Y,ee,ue,null)}}const re=O.children;for(let ee=0,ae=re.length;ee<ae;ee++)D(re[ee],U,Y,ce,E)}function V(O){O.target.removeEventListener("dispose",V);for(const Y in p){const ce=p[Y],E=O.target.uuid;E in ce&&(ce[E].dispose(),delete ce[E])}}}const iw={[Cf]:Rf,[bf]:Df,[Pf]:Nf,[eo]:Lf,[Rf]:Cf,[Df]:bf,[Nf]:Pf,[Lf]:eo};function rw(r){function e(){let G=!1;const Le=new Vt;let oe=null;const me=new Vt(0,0,0,0);return{setMask:function(Re){oe!==Re&&!G&&(r.colorMask(Re,Re,Re,Re),oe=Re)},setLocked:function(Re){G=Re},setClear:function(Re,Ie,ft,kt,ln){ln===!0&&(Re*=kt,Ie*=kt,ft*=kt),Le.set(Re,Ie,ft,kt),me.equals(Le)===!1&&(r.clearColor(Re,Ie,ft,kt),me.copy(Le))},reset:function(){G=!1,oe=null,me.set(-1,0,0,0)}}}function t(){let G=!1,Le=!1,oe=null,me=null,Re=null;return{setReversed:function(Ie){Le=Ie},setTest:function(Ie){Ie?Se(r.DEPTH_TEST):Me(r.DEPTH_TEST)},setMask:function(Ie){oe!==Ie&&!G&&(r.depthMask(Ie),oe=Ie)},setFunc:function(Ie){if(Le&&(Ie=iw[Ie]),me!==Ie){switch(Ie){case Cf:r.depthFunc(r.NEVER);break;case Rf:r.depthFunc(r.ALWAYS);break;case bf:r.depthFunc(r.LESS);break;case eo:r.depthFunc(r.LEQUAL);break;case Pf:r.depthFunc(r.EQUAL);break;case Lf:r.depthFunc(r.GEQUAL);break;case Df:r.depthFunc(r.GREATER);break;case Nf:r.depthFunc(r.NOTEQUAL);break;default:r.depthFunc(r.LEQUAL)}me=Ie}},setLocked:function(Ie){G=Ie},setClear:function(Ie){Re!==Ie&&(r.clearDepth(Ie),Re=Ie)},reset:function(){G=!1,oe=null,me=null,Re=null}}}function s(){let G=!1,Le=null,oe=null,me=null,Re=null,Ie=null,ft=null,kt=null,ln=null;return{setTest:function(mt){G||(mt?Se(r.STENCIL_TEST):Me(r.STENCIL_TEST))},setMask:function(mt){Le!==mt&&!G&&(r.stencilMask(mt),Le=mt)},setFunc:function(mt,en,Gn){(oe!==mt||me!==en||Re!==Gn)&&(r.stencilFunc(mt,en,Gn),oe=mt,me=en,Re=Gn)},setOp:function(mt,en,Gn){(Ie!==mt||ft!==en||kt!==Gn)&&(r.stencilOp(mt,en,Gn),Ie=mt,ft=en,kt=Gn)},setLocked:function(mt){G=mt},setClear:function(mt){ln!==mt&&(r.clearStencil(mt),ln=mt)},reset:function(){G=!1,Le=null,oe=null,me=null,Re=null,Ie=null,ft=null,kt=null,ln=null}}}const a=new e,l=new t,u=new s,f=new WeakMap,h=new WeakMap;let p={},m={},_=new WeakMap,x=[],S=null,T=!1,w=null,v=null,y=null,P=null,b=null,D=null,V=null,O=new vt(0,0,0),U=0,Y=!1,ce=null,E=null,C=null,re=null,ee=null;const ae=r.getParameter(r.MAX_COMBINED_TEXTURE_IMAGE_UNITS);let ue=!1,Z=0;const le=r.getParameter(r.VERSION);le.indexOf("WebGL")!==-1?(Z=parseFloat(/^WebGL (\d)/.exec(le)[1]),ue=Z>=1):le.indexOf("OpenGL ES")!==-1&&(Z=parseFloat(/^OpenGL ES (\d)/.exec(le)[1]),ue=Z>=2);let F=null,se={};const L=r.getParameter(r.SCISSOR_BOX),X=r.getParameter(r.VIEWPORT),ve=new Vt().fromArray(L),Ne=new Vt().fromArray(X);function J(G,Le,oe,me){const Re=new Uint8Array(4),Ie=r.createTexture();r.bindTexture(G,Ie),r.texParameteri(G,r.TEXTURE_MIN_FILTER,r.NEAREST),r.texParameteri(G,r.TEXTURE_MAG_FILTER,r.NEAREST);for(let ft=0;ft<oe;ft++)G===r.TEXTURE_3D||G===r.TEXTURE_2D_ARRAY?r.texImage3D(Le,0,r.RGBA,1,1,me,0,r.RGBA,r.UNSIGNED_BYTE,Re):r.texImage2D(Le+ft,0,r.RGBA,1,1,0,r.RGBA,r.UNSIGNED_BYTE,Re);return Ie}const fe={};fe[r.TEXTURE_2D]=J(r.TEXTURE_2D,r.TEXTURE_2D,1),fe[r.TEXTURE_CUBE_MAP]=J(r.TEXTURE_CUBE_MAP,r.TEXTURE_CUBE_MAP_POSITIVE_X,6),fe[r.TEXTURE_2D_ARRAY]=J(r.TEXTURE_2D_ARRAY,r.TEXTURE_2D_ARRAY,1,1),fe[r.TEXTURE_3D]=J(r.TEXTURE_3D,r.TEXTURE_3D,1,1),a.setClear(0,0,0,1),l.setClear(1),u.setClear(0),Se(r.DEPTH_TEST),l.setFunc(eo),at(!1),ht(gm),Se(r.CULL_FACE),z(Tr);function Se(G){p[G]!==!0&&(r.enable(G),p[G]=!0)}function Me(G){p[G]!==!1&&(r.disable(G),p[G]=!1)}function Pe(G,Le){return m[G]!==Le?(r.bindFramebuffer(G,Le),m[G]=Le,G===r.DRAW_FRAMEBUFFER&&(m[r.FRAMEBUFFER]=Le),G===r.FRAMEBUFFER&&(m[r.DRAW_FRAMEBUFFER]=Le),!0):!1}function Ge(G,Le){let oe=x,me=!1;if(G){oe=_.get(Le),oe===void 0&&(oe=[],_.set(Le,oe));const Re=G.textures;if(oe.length!==Re.length||oe[0]!==r.COLOR_ATTACHMENT0){for(let Ie=0,ft=Re.length;Ie<ft;Ie++)oe[Ie]=r.COLOR_ATTACHMENT0+Ie;oe.length=Re.length,me=!0}}else oe[0]!==r.BACK&&(oe[0]=r.BACK,me=!0);me&&r.drawBuffers(oe)}function dt(G){return S!==G?(r.useProgram(G),S=G,!0):!1}const gt={[Qr]:r.FUNC_ADD,[xx]:r.FUNC_SUBTRACT,[yx]:r.FUNC_REVERSE_SUBTRACT};gt[Sx]=r.MIN,gt[Mx]=r.MAX;const ct={[Ex]:r.ZERO,[Tx]:r.ONE,[wx]:r.SRC_COLOR,[wf]:r.SRC_ALPHA,[Lx]:r.SRC_ALPHA_SATURATE,[bx]:r.DST_COLOR,[Cx]:r.DST_ALPHA,[Ax]:r.ONE_MINUS_SRC_COLOR,[Af]:r.ONE_MINUS_SRC_ALPHA,[Px]:r.ONE_MINUS_DST_COLOR,[Rx]:r.ONE_MINUS_DST_ALPHA,[Dx]:r.CONSTANT_COLOR,[Nx]:r.ONE_MINUS_CONSTANT_COLOR,[Ix]:r.CONSTANT_ALPHA,[Ux]:r.ONE_MINUS_CONSTANT_ALPHA};function z(G,Le,oe,me,Re,Ie,ft,kt,ln,mt){if(G===Tr){T===!0&&(Me(r.BLEND),T=!1);return}if(T===!1&&(Se(r.BLEND),T=!0),G!==vx){if(G!==w||mt!==Y){if((v!==Qr||b!==Qr)&&(r.blendEquation(r.FUNC_ADD),v=Qr,b=Qr),mt)switch(G){case Zs:r.blendFuncSeparate(r.ONE,r.ONE_MINUS_SRC_ALPHA,r.ONE,r.ONE_MINUS_SRC_ALPHA);break;case _m:r.blendFunc(r.ONE,r.ONE);break;case vm:r.blendFuncSeparate(r.ZERO,r.ONE_MINUS_SRC_COLOR,r.ZERO,r.ONE);break;case xm:r.blendFuncSeparate(r.ZERO,r.SRC_COLOR,r.ZERO,r.SRC_ALPHA);break;default:console.error("THREE.WebGLState: Invalid blending: ",G);break}else switch(G){case Zs:r.blendFuncSeparate(r.SRC_ALPHA,r.ONE_MINUS_SRC_ALPHA,r.ONE,r.ONE_MINUS_SRC_ALPHA);break;case _m:r.blendFunc(r.SRC_ALPHA,r.ONE);break;case vm:r.blendFuncSeparate(r.ZERO,r.ONE_MINUS_SRC_COLOR,r.ZERO,r.ONE);break;case xm:r.blendFunc(r.ZERO,r.SRC_COLOR);break;default:console.error("THREE.WebGLState: Invalid blending: ",G);break}y=null,P=null,D=null,V=null,O.set(0,0,0),U=0,w=G,Y=mt}return}Re=Re||Le,Ie=Ie||oe,ft=ft||me,(Le!==v||Re!==b)&&(r.blendEquationSeparate(gt[Le],gt[Re]),v=Le,b=Re),(oe!==y||me!==P||Ie!==D||ft!==V)&&(r.blendFuncSeparate(ct[oe],ct[me],ct[Ie],ct[ft]),y=oe,P=me,D=Ie,V=ft),(kt.equals(O)===!1||ln!==U)&&(r.blendColor(kt.r,kt.g,kt.b,ln),O.copy(kt),U=ln),w=G,Y=!1}function an(G,Le){G.side===zi?Me(r.CULL_FACE):Se(r.CULL_FACE);let oe=G.side===Ln;Le&&(oe=!oe),at(oe),G.blending===Zs&&G.transparent===!1?z(Tr):z(G.blending,G.blendEquation,G.blendSrc,G.blendDst,G.blendEquationAlpha,G.blendSrcAlpha,G.blendDstAlpha,G.blendColor,G.blendAlpha,G.premultipliedAlpha),l.setFunc(G.depthFunc),l.setTest(G.depthTest),l.setMask(G.depthWrite),a.setMask(G.colorWrite);const me=G.stencilWrite;u.setTest(me),me&&(u.setMask(G.stencilWriteMask),u.setFunc(G.stencilFunc,G.stencilRef,G.stencilFuncMask),u.setOp(G.stencilFail,G.stencilZFail,G.stencilZPass)),At(G.polygonOffset,G.polygonOffsetFactor,G.polygonOffsetUnits),G.alphaToCoverage===!0?Se(r.SAMPLE_ALPHA_TO_COVERAGE):Me(r.SAMPLE_ALPHA_TO_COVERAGE)}function at(G){ce!==G&&(G?r.frontFace(r.CW):r.frontFace(r.CCW),ce=G)}function ht(G){G!==mx?(Se(r.CULL_FACE),G!==E&&(G===gm?r.cullFace(r.BACK):G===gx?r.cullFace(r.FRONT):r.cullFace(r.FRONT_AND_BACK))):Me(r.CULL_FACE),E=G}function Ze(G){G!==C&&(ue&&r.lineWidth(G),C=G)}function At(G,Le,oe){G?(Se(r.POLYGON_OFFSET_FILL),(re!==Le||ee!==oe)&&(r.polygonOffset(Le,oe),re=Le,ee=oe)):Me(r.POLYGON_OFFSET_FILL)}function Je(G){G?Se(r.SCISSOR_TEST):Me(r.SCISSOR_TEST)}function N(G){G===void 0&&(G=r.TEXTURE0+ae-1),F!==G&&(r.activeTexture(G),F=G)}function A(G,Le,oe){oe===void 0&&(F===null?oe=r.TEXTURE0+ae-1:oe=F);let me=se[oe];me===void 0&&(me={type:void 0,texture:void 0},se[oe]=me),(me.type!==G||me.texture!==Le)&&(F!==oe&&(r.activeTexture(oe),F=oe),r.bindTexture(G,Le||fe[G]),me.type=G,me.texture=Le)}function K(){const G=se[F];G!==void 0&&G.type!==void 0&&(r.bindTexture(G.type,null),G.type=void 0,G.texture=void 0)}function pe(){try{r.compressedTexImage2D.apply(r,arguments)}catch(G){console.error("THREE.WebGLState:",G)}}function xe(){try{r.compressedTexImage3D.apply(r,arguments)}catch(G){console.error("THREE.WebGLState:",G)}}function de(){try{r.texSubImage2D.apply(r,arguments)}catch(G){console.error("THREE.WebGLState:",G)}}function Ye(){try{r.texSubImage3D.apply(r,arguments)}catch(G){console.error("THREE.WebGLState:",G)}}function Ce(){try{r.compressedTexSubImage2D.apply(r,arguments)}catch(G){console.error("THREE.WebGLState:",G)}}function Fe(){try{r.compressedTexSubImage3D.apply(r,arguments)}catch(G){console.error("THREE.WebGLState:",G)}}function pt(){try{r.texStorage2D.apply(r,arguments)}catch(G){console.error("THREE.WebGLState:",G)}}function Ee(){try{r.texStorage3D.apply(r,arguments)}catch(G){console.error("THREE.WebGLState:",G)}}function Oe(){try{r.texImage2D.apply(r,arguments)}catch(G){console.error("THREE.WebGLState:",G)}}function tt(){try{r.texImage3D.apply(r,arguments)}catch(G){console.error("THREE.WebGLState:",G)}}function et(G){ve.equals(G)===!1&&(r.scissor(G.x,G.y,G.z,G.w),ve.copy(G))}function Be(G){Ne.equals(G)===!1&&(r.viewport(G.x,G.y,G.z,G.w),Ne.copy(G))}function ut(G,Le){let oe=h.get(Le);oe===void 0&&(oe=new WeakMap,h.set(Le,oe));let me=oe.get(G);me===void 0&&(me=r.getUniformBlockIndex(Le,G.name),oe.set(G,me))}function it(G,Le){const me=h.get(Le).get(G);f.get(Le)!==me&&(r.uniformBlockBinding(Le,me,G.__bindingPointIndex),f.set(Le,me))}function Et(){r.disable(r.BLEND),r.disable(r.CULL_FACE),r.disable(r.DEPTH_TEST),r.disable(r.POLYGON_OFFSET_FILL),r.disable(r.SCISSOR_TEST),r.disable(r.STENCIL_TEST),r.disable(r.SAMPLE_ALPHA_TO_COVERAGE),r.blendEquation(r.FUNC_ADD),r.blendFunc(r.ONE,r.ZERO),r.blendFuncSeparate(r.ONE,r.ZERO,r.ONE,r.ZERO),r.blendColor(0,0,0,0),r.colorMask(!0,!0,!0,!0),r.clearColor(0,0,0,0),r.depthMask(!0),r.depthFunc(r.LESS),r.clearDepth(1),r.stencilMask(4294967295),r.stencilFunc(r.ALWAYS,0,4294967295),r.stencilOp(r.KEEP,r.KEEP,r.KEEP),r.clearStencil(0),r.cullFace(r.BACK),r.frontFace(r.CCW),r.polygonOffset(0,0),r.activeTexture(r.TEXTURE0),r.bindFramebuffer(r.FRAMEBUFFER,null),r.bindFramebuffer(r.DRAW_FRAMEBUFFER,null),r.bindFramebuffer(r.READ_FRAMEBUFFER,null),r.useProgram(null),r.lineWidth(1),r.scissor(0,0,r.canvas.width,r.canvas.height),r.viewport(0,0,r.canvas.width,r.canvas.height),p={},F=null,se={},m={},_=new WeakMap,x=[],S=null,T=!1,w=null,v=null,y=null,P=null,b=null,D=null,V=null,O=new vt(0,0,0),U=0,Y=!1,ce=null,E=null,C=null,re=null,ee=null,ve.set(0,0,r.canvas.width,r.canvas.height),Ne.set(0,0,r.canvas.width,r.canvas.height),a.reset(),l.reset(),u.reset()}return{buffers:{color:a,depth:l,stencil:u},enable:Se,disable:Me,bindFramebuffer:Pe,drawBuffers:Ge,useProgram:dt,setBlending:z,setMaterial:an,setFlipSided:at,setCullFace:ht,setLineWidth:Ze,setPolygonOffset:At,setScissorTest:Je,activeTexture:N,bindTexture:A,unbindTexture:K,compressedTexImage2D:pe,compressedTexImage3D:xe,texImage2D:Oe,texImage3D:tt,updateUBOMapping:ut,uniformBlockBinding:it,texStorage2D:pt,texStorage3D:Ee,texSubImage2D:de,texSubImage3D:Ye,compressedTexSubImage2D:Ce,compressedTexSubImage3D:Fe,scissor:et,viewport:Be,reset:Et}}function cg(r,e,t,s){const a=sw(s);switch(t){case kg:return r*e;case zg:return r*e;case Hg:return r*e*2;case Vg:return r*e/a.components*a.byteLength;case yd:return r*e/a.components*a.byteLength;case Gg:return r*e*2/a.components*a.byteLength;case Sd:return r*e*2/a.components*a.byteLength;case Bg:return r*e*3/a.components*a.byteLength;case di:return r*e*4/a.components*a.byteLength;case Md:return r*e*4/a.components*a.byteLength;case kl:case Bl:return Math.floor((r+3)/4)*Math.floor((e+3)/4)*8;case zl:case Hl:return Math.floor((r+3)/4)*Math.floor((e+3)/4)*16;case Bf:case Hf:return Math.max(r,16)*Math.max(e,8)/4;case kf:case zf:return Math.max(r,8)*Math.max(e,8)/2;case Vf:case Gf:return Math.floor((r+3)/4)*Math.floor((e+3)/4)*8;case Wf:return Math.floor((r+3)/4)*Math.floor((e+3)/4)*16;case jf:return Math.floor((r+3)/4)*Math.floor((e+3)/4)*16;case Xf:return Math.floor((r+4)/5)*Math.floor((e+3)/4)*16;case Yf:return Math.floor((r+4)/5)*Math.floor((e+4)/5)*16;case qf:return Math.floor((r+5)/6)*Math.floor((e+4)/5)*16;case $f:return Math.floor((r+5)/6)*Math.floor((e+5)/6)*16;case Kf:return Math.floor((r+7)/8)*Math.floor((e+4)/5)*16;case Zf:return Math.floor((r+7)/8)*Math.floor((e+5)/6)*16;case Qf:return Math.floor((r+7)/8)*Math.floor((e+7)/8)*16;case Jf:return Math.floor((r+9)/10)*Math.floor((e+4)/5)*16;case ed:return Math.floor((r+9)/10)*Math.floor((e+5)/6)*16;case td:return Math.floor((r+9)/10)*Math.floor((e+7)/8)*16;case nd:return Math.floor((r+9)/10)*Math.floor((e+9)/10)*16;case id:return Math.floor((r+11)/12)*Math.floor((e+9)/10)*16;case rd:return Math.floor((r+11)/12)*Math.floor((e+11)/12)*16;case Vl:case sd:case od:return Math.ceil(r/4)*Math.ceil(e/4)*16;case Wg:case ad:return Math.ceil(r/4)*Math.ceil(e/4)*8;case ld:case cd:return Math.ceil(r/4)*Math.ceil(e/4)*16}throw new Error(`Unable to determine texture byte length for ${t} format.`)}function sw(r){switch(r){case Wi:case Ug:return{byteLength:1,components:1};case na:case Fg:case sa:return{byteLength:2,components:1};case vd:case xd:return{byteLength:2,components:4};case ns:case _d:case Hi:return{byteLength:4,components:1};case Og:return{byteLength:4,components:3}}throw new Error(`Unknown texture type ${r}.`)}function ow(r,e,t,s,a,l,u){const f=e.has("WEBGL_multisampled_render_to_texture")?e.get("WEBGL_multisampled_render_to_texture"):null,h=typeof navigator>"u"?!1:/OculusBrowser/g.test(navigator.userAgent),p=new rt,m=new WeakMap;let _;const x=new WeakMap;let S=!1;try{S=typeof OffscreenCanvas<"u"&&new OffscreenCanvas(1,1).getContext("2d")!==null}catch{}function T(N,A){return S?new OffscreenCanvas(N,A):Zl("canvas")}function w(N,A,K){let pe=1;const xe=Je(N);if((xe.width>K||xe.height>K)&&(pe=K/Math.max(xe.width,xe.height)),pe<1)if(typeof HTMLImageElement<"u"&&N instanceof HTMLImageElement||typeof HTMLCanvasElement<"u"&&N instanceof HTMLCanvasElement||typeof ImageBitmap<"u"&&N instanceof ImageBitmap||typeof VideoFrame<"u"&&N instanceof VideoFrame){const de=Math.floor(pe*xe.width),Ye=Math.floor(pe*xe.height);_===void 0&&(_=T(de,Ye));const Ce=A?T(de,Ye):_;return Ce.width=de,Ce.height=Ye,Ce.getContext("2d").drawImage(N,0,0,de,Ye),console.warn("THREE.WebGLRenderer: Texture has been resized from ("+xe.width+"x"+xe.height+") to ("+de+"x"+Ye+")."),Ce}else return"data"in N&&console.warn("THREE.WebGLRenderer: Image in DataTexture is too big ("+xe.width+"x"+xe.height+")."),N;return N}function v(N){return N.generateMipmaps&&N.minFilter!==Qn&&N.minFilter!==ui}function y(N){r.generateMipmap(N)}function P(N,A,K,pe,xe=!1){if(N!==null){if(r[N]!==void 0)return r[N];console.warn("THREE.WebGLRenderer: Attempt to use non-existing WebGL internal format '"+N+"'")}let de=A;if(A===r.RED&&(K===r.FLOAT&&(de=r.R32F),K===r.HALF_FLOAT&&(de=r.R16F),K===r.UNSIGNED_BYTE&&(de=r.R8)),A===r.RED_INTEGER&&(K===r.UNSIGNED_BYTE&&(de=r.R8UI),K===r.UNSIGNED_SHORT&&(de=r.R16UI),K===r.UNSIGNED_INT&&(de=r.R32UI),K===r.BYTE&&(de=r.R8I),K===r.SHORT&&(de=r.R16I),K===r.INT&&(de=r.R32I)),A===r.RG&&(K===r.FLOAT&&(de=r.RG32F),K===r.HALF_FLOAT&&(de=r.RG16F),K===r.UNSIGNED_BYTE&&(de=r.RG8)),A===r.RG_INTEGER&&(K===r.UNSIGNED_BYTE&&(de=r.RG8UI),K===r.UNSIGNED_SHORT&&(de=r.RG16UI),K===r.UNSIGNED_INT&&(de=r.RG32UI),K===r.BYTE&&(de=r.RG8I),K===r.SHORT&&(de=r.RG16I),K===r.INT&&(de=r.RG32I)),A===r.RGB_INTEGER&&(K===r.UNSIGNED_BYTE&&(de=r.RGB8UI),K===r.UNSIGNED_SHORT&&(de=r.RGB16UI),K===r.UNSIGNED_INT&&(de=r.RGB32UI),K===r.BYTE&&(de=r.RGB8I),K===r.SHORT&&(de=r.RGB16I),K===r.INT&&(de=r.RGB32I)),A===r.RGBA_INTEGER&&(K===r.UNSIGNED_BYTE&&(de=r.RGBA8UI),K===r.UNSIGNED_SHORT&&(de=r.RGBA16UI),K===r.UNSIGNED_INT&&(de=r.RGBA32UI),K===r.BYTE&&(de=r.RGBA8I),K===r.SHORT&&(de=r.RGBA16I),K===r.INT&&(de=r.RGBA32I)),A===r.RGB&&K===r.UNSIGNED_INT_5_9_9_9_REV&&(de=r.RGB9_E5),A===r.RGBA){const Ye=xe?Yl:Tt.getTransfer(pe);K===r.FLOAT&&(de=r.RGBA32F),K===r.HALF_FLOAT&&(de=r.RGBA16F),K===r.UNSIGNED_BYTE&&(de=Ye===Ut?r.SRGB8_ALPHA8:r.RGBA8),K===r.UNSIGNED_SHORT_4_4_4_4&&(de=r.RGBA4),K===r.UNSIGNED_SHORT_5_5_5_1&&(de=r.RGB5_A1)}return(de===r.R16F||de===r.R32F||de===r.RG16F||de===r.RG32F||de===r.RGBA16F||de===r.RGBA32F)&&e.get("EXT_color_buffer_float"),de}function b(N,A){let K;return N?A===null||A===ns||A===io?K=r.DEPTH24_STENCIL8:A===Hi?K=r.DEPTH32F_STENCIL8:A===na&&(K=r.DEPTH24_STENCIL8,console.warn("DepthTexture: 16 bit depth attachment is not supported with stencil. Using 24-bit attachment.")):A===null||A===ns||A===io?K=r.DEPTH_COMPONENT24:A===Hi?K=r.DEPTH_COMPONENT32F:A===na&&(K=r.DEPTH_COMPONENT16),K}function D(N,A){return v(N)===!0||N.isFramebufferTexture&&N.minFilter!==Qn&&N.minFilter!==ui?Math.log2(Math.max(A.width,A.height))+1:N.mipmaps!==void 0&&N.mipmaps.length>0?N.mipmaps.length:N.isCompressedTexture&&Array.isArray(N.image)?A.mipmaps.length:1}function V(N){const A=N.target;A.removeEventListener("dispose",V),U(A),A.isVideoTexture&&m.delete(A)}function O(N){const A=N.target;A.removeEventListener("dispose",O),ce(A)}function U(N){const A=s.get(N);if(A.__webglInit===void 0)return;const K=N.source,pe=x.get(K);if(pe){const xe=pe[A.__cacheKey];xe.usedTimes--,xe.usedTimes===0&&Y(N),Object.keys(pe).length===0&&x.delete(K)}s.remove(N)}function Y(N){const A=s.get(N);r.deleteTexture(A.__webglTexture);const K=N.source,pe=x.get(K);delete pe[A.__cacheKey],u.memory.textures--}function ce(N){const A=s.get(N);if(N.depthTexture&&N.depthTexture.dispose(),N.isWebGLCubeRenderTarget)for(let pe=0;pe<6;pe++){if(Array.isArray(A.__webglFramebuffer[pe]))for(let xe=0;xe<A.__webglFramebuffer[pe].length;xe++)r.deleteFramebuffer(A.__webglFramebuffer[pe][xe]);else r.deleteFramebuffer(A.__webglFramebuffer[pe]);A.__webglDepthbuffer&&r.deleteRenderbuffer(A.__webglDepthbuffer[pe])}else{if(Array.isArray(A.__webglFramebuffer))for(let pe=0;pe<A.__webglFramebuffer.length;pe++)r.deleteFramebuffer(A.__webglFramebuffer[pe]);else r.deleteFramebuffer(A.__webglFramebuffer);if(A.__webglDepthbuffer&&r.deleteRenderbuffer(A.__webglDepthbuffer),A.__webglMultisampledFramebuffer&&r.deleteFramebuffer(A.__webglMultisampledFramebuffer),A.__webglColorRenderbuffer)for(let pe=0;pe<A.__webglColorRenderbuffer.length;pe++)A.__webglColorRenderbuffer[pe]&&r.deleteRenderbuffer(A.__webglColorRenderbuffer[pe]);A.__webglDepthRenderbuffer&&r.deleteRenderbuffer(A.__webglDepthRenderbuffer)}const K=N.textures;for(let pe=0,xe=K.length;pe<xe;pe++){const de=s.get(K[pe]);de.__webglTexture&&(r.deleteTexture(de.__webglTexture),u.memory.textures--),s.remove(K[pe])}s.remove(N)}let E=0;function C(){E=0}function re(){const N=E;return N>=a.maxTextures&&console.warn("THREE.WebGLTextures: Trying to use "+N+" texture units while this GPU supports only "+a.maxTextures),E+=1,N}function ee(N){const A=[];return A.push(N.wrapS),A.push(N.wrapT),A.push(N.wrapR||0),A.push(N.magFilter),A.push(N.minFilter),A.push(N.anisotropy),A.push(N.internalFormat),A.push(N.format),A.push(N.type),A.push(N.generateMipmaps),A.push(N.premultiplyAlpha),A.push(N.flipY),A.push(N.unpackAlignment),A.push(N.colorSpace),A.join()}function ae(N,A){const K=s.get(N);if(N.isVideoTexture&&Ze(N),N.isRenderTargetTexture===!1&&N.version>0&&K.__version!==N.version){const pe=N.image;if(pe===null)console.warn("THREE.WebGLRenderer: Texture marked for update but no image data found.");else if(pe.complete===!1)console.warn("THREE.WebGLRenderer: Texture marked for update but image is incomplete");else{Ne(K,N,A);return}}t.bindTexture(r.TEXTURE_2D,K.__webglTexture,r.TEXTURE0+A)}function ue(N,A){const K=s.get(N);if(N.version>0&&K.__version!==N.version){Ne(K,N,A);return}t.bindTexture(r.TEXTURE_2D_ARRAY,K.__webglTexture,r.TEXTURE0+A)}function Z(N,A){const K=s.get(N);if(N.version>0&&K.__version!==N.version){Ne(K,N,A);return}t.bindTexture(r.TEXTURE_3D,K.__webglTexture,r.TEXTURE0+A)}function le(N,A){const K=s.get(N);if(N.version>0&&K.__version!==N.version){J(K,N,A);return}t.bindTexture(r.TEXTURE_CUBE_MAP,K.__webglTexture,r.TEXTURE0+A)}const F={[Ff]:r.REPEAT,[es]:r.CLAMP_TO_EDGE,[Of]:r.MIRRORED_REPEAT},se={[Qn]:r.NEAREST,[jx]:r.NEAREST_MIPMAP_NEAREST,[ml]:r.NEAREST_MIPMAP_LINEAR,[ui]:r.LINEAR,[Vu]:r.LINEAR_MIPMAP_NEAREST,[ts]:r.LINEAR_MIPMAP_LINEAR},L={[$x]:r.NEVER,[ty]:r.ALWAYS,[Kx]:r.LESS,[Xg]:r.LEQUAL,[Zx]:r.EQUAL,[ey]:r.GEQUAL,[Qx]:r.GREATER,[Jx]:r.NOTEQUAL};function X(N,A){if(A.type===Hi&&e.has("OES_texture_float_linear")===!1&&(A.magFilter===ui||A.magFilter===Vu||A.magFilter===ml||A.magFilter===ts||A.minFilter===ui||A.minFilter===Vu||A.minFilter===ml||A.minFilter===ts)&&console.warn("THREE.WebGLRenderer: Unable to use linear filtering with floating point textures. OES_texture_float_linear not supported on this device."),r.texParameteri(N,r.TEXTURE_WRAP_S,F[A.wrapS]),r.texParameteri(N,r.TEXTURE_WRAP_T,F[A.wrapT]),(N===r.TEXTURE_3D||N===r.TEXTURE_2D_ARRAY)&&r.texParameteri(N,r.TEXTURE_WRAP_R,F[A.wrapR]),r.texParameteri(N,r.TEXTURE_MAG_FILTER,se[A.magFilter]),r.texParameteri(N,r.TEXTURE_MIN_FILTER,se[A.minFilter]),A.compareFunction&&(r.texParameteri(N,r.TEXTURE_COMPARE_MODE,r.COMPARE_REF_TO_TEXTURE),r.texParameteri(N,r.TEXTURE_COMPARE_FUNC,L[A.compareFunction])),e.has("EXT_texture_filter_anisotropic")===!0){if(A.magFilter===Qn||A.minFilter!==ml&&A.minFilter!==ts||A.type===Hi&&e.has("OES_texture_float_linear")===!1)return;if(A.anisotropy>1||s.get(A).__currentAnisotropy){const K=e.get("EXT_texture_filter_anisotropic");r.texParameterf(N,K.TEXTURE_MAX_ANISOTROPY_EXT,Math.min(A.anisotropy,a.getMaxAnisotropy())),s.get(A).__currentAnisotropy=A.anisotropy}}}function ve(N,A){let K=!1;N.__webglInit===void 0&&(N.__webglInit=!0,A.addEventListener("dispose",V));const pe=A.source;let xe=x.get(pe);xe===void 0&&(xe={},x.set(pe,xe));const de=ee(A);if(de!==N.__cacheKey){xe[de]===void 0&&(xe[de]={texture:r.createTexture(),usedTimes:0},u.memory.textures++,K=!0),xe[de].usedTimes++;const Ye=xe[N.__cacheKey];Ye!==void 0&&(xe[N.__cacheKey].usedTimes--,Ye.usedTimes===0&&Y(A)),N.__cacheKey=de,N.__webglTexture=xe[de].texture}return K}function Ne(N,A,K){let pe=r.TEXTURE_2D;(A.isDataArrayTexture||A.isCompressedArrayTexture)&&(pe=r.TEXTURE_2D_ARRAY),A.isData3DTexture&&(pe=r.TEXTURE_3D);const xe=ve(N,A),de=A.source;t.bindTexture(pe,N.__webglTexture,r.TEXTURE0+K);const Ye=s.get(de);if(de.version!==Ye.__version||xe===!0){t.activeTexture(r.TEXTURE0+K);const Ce=Tt.getPrimaries(Tt.workingColorSpace),Fe=A.colorSpace===Sr?null:Tt.getPrimaries(A.colorSpace),pt=A.colorSpace===Sr||Ce===Fe?r.NONE:r.BROWSER_DEFAULT_WEBGL;r.pixelStorei(r.UNPACK_FLIP_Y_WEBGL,A.flipY),r.pixelStorei(r.UNPACK_PREMULTIPLY_ALPHA_WEBGL,A.premultiplyAlpha),r.pixelStorei(r.UNPACK_ALIGNMENT,A.unpackAlignment),r.pixelStorei(r.UNPACK_COLORSPACE_CONVERSION_WEBGL,pt);let Ee=w(A.image,!1,a.maxTextureSize);Ee=At(A,Ee);const Oe=l.convert(A.format,A.colorSpace),tt=l.convert(A.type);let et=P(A.internalFormat,Oe,tt,A.colorSpace,A.isVideoTexture);X(pe,A);let Be;const ut=A.mipmaps,it=A.isVideoTexture!==!0,Et=Ye.__version===void 0||xe===!0,G=de.dataReady,Le=D(A,Ee);if(A.isDepthTexture)et=b(A.format===ro,A.type),Et&&(it?t.texStorage2D(r.TEXTURE_2D,1,et,Ee.width,Ee.height):t.texImage2D(r.TEXTURE_2D,0,et,Ee.width,Ee.height,0,Oe,tt,null));else if(A.isDataTexture)if(ut.length>0){it&&Et&&t.texStorage2D(r.TEXTURE_2D,Le,et,ut[0].width,ut[0].height);for(let oe=0,me=ut.length;oe<me;oe++)Be=ut[oe],it?G&&t.texSubImage2D(r.TEXTURE_2D,oe,0,0,Be.width,Be.height,Oe,tt,Be.data):t.texImage2D(r.TEXTURE_2D,oe,et,Be.width,Be.height,0,Oe,tt,Be.data);A.generateMipmaps=!1}else it?(Et&&t.texStorage2D(r.TEXTURE_2D,Le,et,Ee.width,Ee.height),G&&t.texSubImage2D(r.TEXTURE_2D,0,0,0,Ee.width,Ee.height,Oe,tt,Ee.data)):t.texImage2D(r.TEXTURE_2D,0,et,Ee.width,Ee.height,0,Oe,tt,Ee.data);else if(A.isCompressedTexture)if(A.isCompressedArrayTexture){it&&Et&&t.texStorage3D(r.TEXTURE_2D_ARRAY,Le,et,ut[0].width,ut[0].height,Ee.depth);for(let oe=0,me=ut.length;oe<me;oe++)if(Be=ut[oe],A.format!==di)if(Oe!==null)if(it){if(G)if(A.layerUpdates.size>0){const Re=cg(Be.width,Be.height,A.format,A.type);for(const Ie of A.layerUpdates){const ft=Be.data.subarray(Ie*Re/Be.data.BYTES_PER_ELEMENT,(Ie+1)*Re/Be.data.BYTES_PER_ELEMENT);t.compressedTexSubImage3D(r.TEXTURE_2D_ARRAY,oe,0,0,Ie,Be.width,Be.height,1,Oe,ft,0,0)}A.clearLayerUpdates()}else t.compressedTexSubImage3D(r.TEXTURE_2D_ARRAY,oe,0,0,0,Be.width,Be.height,Ee.depth,Oe,Be.data,0,0)}else t.compressedTexImage3D(r.TEXTURE_2D_ARRAY,oe,et,Be.width,Be.height,Ee.depth,0,Be.data,0,0);else console.warn("THREE.WebGLRenderer: Attempt to load unsupported compressed texture format in .uploadTexture()");else it?G&&t.texSubImage3D(r.TEXTURE_2D_ARRAY,oe,0,0,0,Be.width,Be.height,Ee.depth,Oe,tt,Be.data):t.texImage3D(r.TEXTURE_2D_ARRAY,oe,et,Be.width,Be.height,Ee.depth,0,Oe,tt,Be.data)}else{it&&Et&&t.texStorage2D(r.TEXTURE_2D,Le,et,ut[0].width,ut[0].height);for(let oe=0,me=ut.length;oe<me;oe++)Be=ut[oe],A.format!==di?Oe!==null?it?G&&t.compressedTexSubImage2D(r.TEXTURE_2D,oe,0,0,Be.width,Be.height,Oe,Be.data):t.compressedTexImage2D(r.TEXTURE_2D,oe,et,Be.width,Be.height,0,Be.data):console.warn("THREE.WebGLRenderer: Attempt to load unsupported compressed texture format in .uploadTexture()"):it?G&&t.texSubImage2D(r.TEXTURE_2D,oe,0,0,Be.width,Be.height,Oe,tt,Be.data):t.texImage2D(r.TEXTURE_2D,oe,et,Be.width,Be.height,0,Oe,tt,Be.data)}else if(A.isDataArrayTexture)if(it){if(Et&&t.texStorage3D(r.TEXTURE_2D_ARRAY,Le,et,Ee.width,Ee.height,Ee.depth),G)if(A.layerUpdates.size>0){const oe=cg(Ee.width,Ee.height,A.format,A.type);for(const me of A.layerUpdates){const Re=Ee.data.subarray(me*oe/Ee.data.BYTES_PER_ELEMENT,(me+1)*oe/Ee.data.BYTES_PER_ELEMENT);t.texSubImage3D(r.TEXTURE_2D_ARRAY,0,0,0,me,Ee.width,Ee.height,1,Oe,tt,Re)}A.clearLayerUpdates()}else t.texSubImage3D(r.TEXTURE_2D_ARRAY,0,0,0,0,Ee.width,Ee.height,Ee.depth,Oe,tt,Ee.data)}else t.texImage3D(r.TEXTURE_2D_ARRAY,0,et,Ee.width,Ee.height,Ee.depth,0,Oe,tt,Ee.data);else if(A.isData3DTexture)it?(Et&&t.texStorage3D(r.TEXTURE_3D,Le,et,Ee.width,Ee.height,Ee.depth),G&&t.texSubImage3D(r.TEXTURE_3D,0,0,0,0,Ee.width,Ee.height,Ee.depth,Oe,tt,Ee.data)):t.texImage3D(r.TEXTURE_3D,0,et,Ee.width,Ee.height,Ee.depth,0,Oe,tt,Ee.data);else if(A.isFramebufferTexture){if(Et)if(it)t.texStorage2D(r.TEXTURE_2D,Le,et,Ee.width,Ee.height);else{let oe=Ee.width,me=Ee.height;for(let Re=0;Re<Le;Re++)t.texImage2D(r.TEXTURE_2D,Re,et,oe,me,0,Oe,tt,null),oe>>=1,me>>=1}}else if(ut.length>0){if(it&&Et){const oe=Je(ut[0]);t.texStorage2D(r.TEXTURE_2D,Le,et,oe.width,oe.height)}for(let oe=0,me=ut.length;oe<me;oe++)Be=ut[oe],it?G&&t.texSubImage2D(r.TEXTURE_2D,oe,0,0,Oe,tt,Be):t.texImage2D(r.TEXTURE_2D,oe,et,Oe,tt,Be);A.generateMipmaps=!1}else if(it){if(Et){const oe=Je(Ee);t.texStorage2D(r.TEXTURE_2D,Le,et,oe.width,oe.height)}G&&t.texSubImage2D(r.TEXTURE_2D,0,0,0,Oe,tt,Ee)}else t.texImage2D(r.TEXTURE_2D,0,et,Oe,tt,Ee);v(A)&&y(pe),Ye.__version=de.version,A.onUpdate&&A.onUpdate(A)}N.__version=A.version}function J(N,A,K){if(A.image.length!==6)return;const pe=ve(N,A),xe=A.source;t.bindTexture(r.TEXTURE_CUBE_MAP,N.__webglTexture,r.TEXTURE0+K);const de=s.get(xe);if(xe.version!==de.__version||pe===!0){t.activeTexture(r.TEXTURE0+K);const Ye=Tt.getPrimaries(Tt.workingColorSpace),Ce=A.colorSpace===Sr?null:Tt.getPrimaries(A.colorSpace),Fe=A.colorSpace===Sr||Ye===Ce?r.NONE:r.BROWSER_DEFAULT_WEBGL;r.pixelStorei(r.UNPACK_FLIP_Y_WEBGL,A.flipY),r.pixelStorei(r.UNPACK_PREMULTIPLY_ALPHA_WEBGL,A.premultiplyAlpha),r.pixelStorei(r.UNPACK_ALIGNMENT,A.unpackAlignment),r.pixelStorei(r.UNPACK_COLORSPACE_CONVERSION_WEBGL,Fe);const pt=A.isCompressedTexture||A.image[0].isCompressedTexture,Ee=A.image[0]&&A.image[0].isDataTexture,Oe=[];for(let me=0;me<6;me++)!pt&&!Ee?Oe[me]=w(A.image[me],!0,a.maxCubemapSize):Oe[me]=Ee?A.image[me].image:A.image[me],Oe[me]=At(A,Oe[me]);const tt=Oe[0],et=l.convert(A.format,A.colorSpace),Be=l.convert(A.type),ut=P(A.internalFormat,et,Be,A.colorSpace),it=A.isVideoTexture!==!0,Et=de.__version===void 0||pe===!0,G=xe.dataReady;let Le=D(A,tt);X(r.TEXTURE_CUBE_MAP,A);let oe;if(pt){it&&Et&&t.texStorage2D(r.TEXTURE_CUBE_MAP,Le,ut,tt.width,tt.height);for(let me=0;me<6;me++){oe=Oe[me].mipmaps;for(let Re=0;Re<oe.length;Re++){const Ie=oe[Re];A.format!==di?et!==null?it?G&&t.compressedTexSubImage2D(r.TEXTURE_CUBE_MAP_POSITIVE_X+me,Re,0,0,Ie.width,Ie.height,et,Ie.data):t.compressedTexImage2D(r.TEXTURE_CUBE_MAP_POSITIVE_X+me,Re,ut,Ie.width,Ie.height,0,Ie.data):console.warn("THREE.WebGLRenderer: Attempt to load unsupported compressed texture format in .setTextureCube()"):it?G&&t.texSubImage2D(r.TEXTURE_CUBE_MAP_POSITIVE_X+me,Re,0,0,Ie.width,Ie.height,et,Be,Ie.data):t.texImage2D(r.TEXTURE_CUBE_MAP_POSITIVE_X+me,Re,ut,Ie.width,Ie.height,0,et,Be,Ie.data)}}}else{if(oe=A.mipmaps,it&&Et){oe.length>0&&Le++;const me=Je(Oe[0]);t.texStorage2D(r.TEXTURE_CUBE_MAP,Le,ut,me.width,me.height)}for(let me=0;me<6;me++)if(Ee){it?G&&t.texSubImage2D(r.TEXTURE_CUBE_MAP_POSITIVE_X+me,0,0,0,Oe[me].width,Oe[me].height,et,Be,Oe[me].data):t.texImage2D(r.TEXTURE_CUBE_MAP_POSITIVE_X+me,0,ut,Oe[me].width,Oe[me].height,0,et,Be,Oe[me].data);for(let Re=0;Re<oe.length;Re++){const ft=oe[Re].image[me].image;it?G&&t.texSubImage2D(r.TEXTURE_CUBE_MAP_POSITIVE_X+me,Re+1,0,0,ft.width,ft.height,et,Be,ft.data):t.texImage2D(r.TEXTURE_CUBE_MAP_POSITIVE_X+me,Re+1,ut,ft.width,ft.height,0,et,Be,ft.data)}}else{it?G&&t.texSubImage2D(r.TEXTURE_CUBE_MAP_POSITIVE_X+me,0,0,0,et,Be,Oe[me]):t.texImage2D(r.TEXTURE_CUBE_MAP_POSITIVE_X+me,0,ut,et,Be,Oe[me]);for(let Re=0;Re<oe.length;Re++){const Ie=oe[Re];it?G&&t.texSubImage2D(r.TEXTURE_CUBE_MAP_POSITIVE_X+me,Re+1,0,0,et,Be,Ie.image[me]):t.texImage2D(r.TEXTURE_CUBE_MAP_POSITIVE_X+me,Re+1,ut,et,Be,Ie.image[me])}}}v(A)&&y(r.TEXTURE_CUBE_MAP),de.__version=xe.version,A.onUpdate&&A.onUpdate(A)}N.__version=A.version}function fe(N,A,K,pe,xe,de){const Ye=l.convert(K.format,K.colorSpace),Ce=l.convert(K.type),Fe=P(K.internalFormat,Ye,Ce,K.colorSpace);if(!s.get(A).__hasExternalTextures){const Ee=Math.max(1,A.width>>de),Oe=Math.max(1,A.height>>de);xe===r.TEXTURE_3D||xe===r.TEXTURE_2D_ARRAY?t.texImage3D(xe,de,Fe,Ee,Oe,A.depth,0,Ye,Ce,null):t.texImage2D(xe,de,Fe,Ee,Oe,0,Ye,Ce,null)}t.bindFramebuffer(r.FRAMEBUFFER,N),ht(A)?f.framebufferTexture2DMultisampleEXT(r.FRAMEBUFFER,pe,xe,s.get(K).__webglTexture,0,at(A)):(xe===r.TEXTURE_2D||xe>=r.TEXTURE_CUBE_MAP_POSITIVE_X&&xe<=r.TEXTURE_CUBE_MAP_NEGATIVE_Z)&&r.framebufferTexture2D(r.FRAMEBUFFER,pe,xe,s.get(K).__webglTexture,de),t.bindFramebuffer(r.FRAMEBUFFER,null)}function Se(N,A,K){if(r.bindRenderbuffer(r.RENDERBUFFER,N),A.depthBuffer){const pe=A.depthTexture,xe=pe&&pe.isDepthTexture?pe.type:null,de=b(A.stencilBuffer,xe),Ye=A.stencilBuffer?r.DEPTH_STENCIL_ATTACHMENT:r.DEPTH_ATTACHMENT,Ce=at(A);ht(A)?f.renderbufferStorageMultisampleEXT(r.RENDERBUFFER,Ce,de,A.width,A.height):K?r.renderbufferStorageMultisample(r.RENDERBUFFER,Ce,de,A.width,A.height):r.renderbufferStorage(r.RENDERBUFFER,de,A.width,A.height),r.framebufferRenderbuffer(r.FRAMEBUFFER,Ye,r.RENDERBUFFER,N)}else{const pe=A.textures;for(let xe=0;xe<pe.length;xe++){const de=pe[xe],Ye=l.convert(de.format,de.colorSpace),Ce=l.convert(de.type),Fe=P(de.internalFormat,Ye,Ce,de.colorSpace),pt=at(A);K&&ht(A)===!1?r.renderbufferStorageMultisample(r.RENDERBUFFER,pt,Fe,A.width,A.height):ht(A)?f.renderbufferStorageMultisampleEXT(r.RENDERBUFFER,pt,Fe,A.width,A.height):r.renderbufferStorage(r.RENDERBUFFER,Fe,A.width,A.height)}}r.bindRenderbuffer(r.RENDERBUFFER,null)}function Me(N,A){if(A&&A.isWebGLCubeRenderTarget)throw new Error("Depth Texture with cube render targets is not supported");if(t.bindFramebuffer(r.FRAMEBUFFER,N),!(A.depthTexture&&A.depthTexture.isDepthTexture))throw new Error("renderTarget.depthTexture must be an instance of THREE.DepthTexture");(!s.get(A.depthTexture).__webglTexture||A.depthTexture.image.width!==A.width||A.depthTexture.image.height!==A.height)&&(A.depthTexture.image.width=A.width,A.depthTexture.image.height=A.height,A.depthTexture.needsUpdate=!0),ae(A.depthTexture,0);const pe=s.get(A.depthTexture).__webglTexture,xe=at(A);if(A.depthTexture.format===Qs)ht(A)?f.framebufferTexture2DMultisampleEXT(r.FRAMEBUFFER,r.DEPTH_ATTACHMENT,r.TEXTURE_2D,pe,0,xe):r.framebufferTexture2D(r.FRAMEBUFFER,r.DEPTH_ATTACHMENT,r.TEXTURE_2D,pe,0);else if(A.depthTexture.format===ro)ht(A)?f.framebufferTexture2DMultisampleEXT(r.FRAMEBUFFER,r.DEPTH_STENCIL_ATTACHMENT,r.TEXTURE_2D,pe,0,xe):r.framebufferTexture2D(r.FRAMEBUFFER,r.DEPTH_STENCIL_ATTACHMENT,r.TEXTURE_2D,pe,0);else throw new Error("Unknown depthTexture format")}function Pe(N){const A=s.get(N),K=N.isWebGLCubeRenderTarget===!0;if(A.__boundDepthTexture!==N.depthTexture){const pe=N.depthTexture;if(A.__depthDisposeCallback&&A.__depthDisposeCallback(),pe){const xe=()=>{delete A.__boundDepthTexture,delete A.__depthDisposeCallback,pe.removeEventListener("dispose",xe)};pe.addEventListener("dispose",xe),A.__depthDisposeCallback=xe}A.__boundDepthTexture=pe}if(N.depthTexture&&!A.__autoAllocateDepthBuffer){if(K)throw new Error("target.depthTexture not supported in Cube render targets");Me(A.__webglFramebuffer,N)}else if(K){A.__webglDepthbuffer=[];for(let pe=0;pe<6;pe++)if(t.bindFramebuffer(r.FRAMEBUFFER,A.__webglFramebuffer[pe]),A.__webglDepthbuffer[pe]===void 0)A.__webglDepthbuffer[pe]=r.createRenderbuffer(),Se(A.__webglDepthbuffer[pe],N,!1);else{const xe=N.stencilBuffer?r.DEPTH_STENCIL_ATTACHMENT:r.DEPTH_ATTACHMENT,de=A.__webglDepthbuffer[pe];r.bindRenderbuffer(r.RENDERBUFFER,de),r.framebufferRenderbuffer(r.FRAMEBUFFER,xe,r.RENDERBUFFER,de)}}else if(t.bindFramebuffer(r.FRAMEBUFFER,A.__webglFramebuffer),A.__webglDepthbuffer===void 0)A.__webglDepthbuffer=r.createRenderbuffer(),Se(A.__webglDepthbuffer,N,!1);else{const pe=N.stencilBuffer?r.DEPTH_STENCIL_ATTACHMENT:r.DEPTH_ATTACHMENT,xe=A.__webglDepthbuffer;r.bindRenderbuffer(r.RENDERBUFFER,xe),r.framebufferRenderbuffer(r.FRAMEBUFFER,pe,r.RENDERBUFFER,xe)}t.bindFramebuffer(r.FRAMEBUFFER,null)}function Ge(N,A,K){const pe=s.get(N);A!==void 0&&fe(pe.__webglFramebuffer,N,N.texture,r.COLOR_ATTACHMENT0,r.TEXTURE_2D,0),K!==void 0&&Pe(N)}function dt(N){const A=N.texture,K=s.get(N),pe=s.get(A);N.addEventListener("dispose",O);const xe=N.textures,de=N.isWebGLCubeRenderTarget===!0,Ye=xe.length>1;if(Ye||(pe.__webglTexture===void 0&&(pe.__webglTexture=r.createTexture()),pe.__version=A.version,u.memory.textures++),de){K.__webglFramebuffer=[];for(let Ce=0;Ce<6;Ce++)if(A.mipmaps&&A.mipmaps.length>0){K.__webglFramebuffer[Ce]=[];for(let Fe=0;Fe<A.mipmaps.length;Fe++)K.__webglFramebuffer[Ce][Fe]=r.createFramebuffer()}else K.__webglFramebuffer[Ce]=r.createFramebuffer()}else{if(A.mipmaps&&A.mipmaps.length>0){K.__webglFramebuffer=[];for(let Ce=0;Ce<A.mipmaps.length;Ce++)K.__webglFramebuffer[Ce]=r.createFramebuffer()}else K.__webglFramebuffer=r.createFramebuffer();if(Ye)for(let Ce=0,Fe=xe.length;Ce<Fe;Ce++){const pt=s.get(xe[Ce]);pt.__webglTexture===void 0&&(pt.__webglTexture=r.createTexture(),u.memory.textures++)}if(N.samples>0&&ht(N)===!1){K.__webglMultisampledFramebuffer=r.createFramebuffer(),K.__webglColorRenderbuffer=[],t.bindFramebuffer(r.FRAMEBUFFER,K.__webglMultisampledFramebuffer);for(let Ce=0;Ce<xe.length;Ce++){const Fe=xe[Ce];K.__webglColorRenderbuffer[Ce]=r.createRenderbuffer(),r.bindRenderbuffer(r.RENDERBUFFER,K.__webglColorRenderbuffer[Ce]);const pt=l.convert(Fe.format,Fe.colorSpace),Ee=l.convert(Fe.type),Oe=P(Fe.internalFormat,pt,Ee,Fe.colorSpace,N.isXRRenderTarget===!0),tt=at(N);r.renderbufferStorageMultisample(r.RENDERBUFFER,tt,Oe,N.width,N.height),r.framebufferRenderbuffer(r.FRAMEBUFFER,r.COLOR_ATTACHMENT0+Ce,r.RENDERBUFFER,K.__webglColorRenderbuffer[Ce])}r.bindRenderbuffer(r.RENDERBUFFER,null),N.depthBuffer&&(K.__webglDepthRenderbuffer=r.createRenderbuffer(),Se(K.__webglDepthRenderbuffer,N,!0)),t.bindFramebuffer(r.FRAMEBUFFER,null)}}if(de){t.bindTexture(r.TEXTURE_CUBE_MAP,pe.__webglTexture),X(r.TEXTURE_CUBE_MAP,A);for(let Ce=0;Ce<6;Ce++)if(A.mipmaps&&A.mipmaps.length>0)for(let Fe=0;Fe<A.mipmaps.length;Fe++)fe(K.__webglFramebuffer[Ce][Fe],N,A,r.COLOR_ATTACHMENT0,r.TEXTURE_CUBE_MAP_POSITIVE_X+Ce,Fe);else fe(K.__webglFramebuffer[Ce],N,A,r.COLOR_ATTACHMENT0,r.TEXTURE_CUBE_MAP_POSITIVE_X+Ce,0);v(A)&&y(r.TEXTURE_CUBE_MAP),t.unbindTexture()}else if(Ye){for(let Ce=0,Fe=xe.length;Ce<Fe;Ce++){const pt=xe[Ce],Ee=s.get(pt);t.bindTexture(r.TEXTURE_2D,Ee.__webglTexture),X(r.TEXTURE_2D,pt),fe(K.__webglFramebuffer,N,pt,r.COLOR_ATTACHMENT0+Ce,r.TEXTURE_2D,0),v(pt)&&y(r.TEXTURE_2D)}t.unbindTexture()}else{let Ce=r.TEXTURE_2D;if((N.isWebGL3DRenderTarget||N.isWebGLArrayRenderTarget)&&(Ce=N.isWebGL3DRenderTarget?r.TEXTURE_3D:r.TEXTURE_2D_ARRAY),t.bindTexture(Ce,pe.__webglTexture),X(Ce,A),A.mipmaps&&A.mipmaps.length>0)for(let Fe=0;Fe<A.mipmaps.length;Fe++)fe(K.__webglFramebuffer[Fe],N,A,r.COLOR_ATTACHMENT0,Ce,Fe);else fe(K.__webglFramebuffer,N,A,r.COLOR_ATTACHMENT0,Ce,0);v(A)&&y(Ce),t.unbindTexture()}N.depthBuffer&&Pe(N)}function gt(N){const A=N.textures;for(let K=0,pe=A.length;K<pe;K++){const xe=A[K];if(v(xe)){const de=N.isWebGLCubeRenderTarget?r.TEXTURE_CUBE_MAP:r.TEXTURE_2D,Ye=s.get(xe).__webglTexture;t.bindTexture(de,Ye),y(de),t.unbindTexture()}}}const ct=[],z=[];function an(N){if(N.samples>0){if(ht(N)===!1){const A=N.textures,K=N.width,pe=N.height;let xe=r.COLOR_BUFFER_BIT;const de=N.stencilBuffer?r.DEPTH_STENCIL_ATTACHMENT:r.DEPTH_ATTACHMENT,Ye=s.get(N),Ce=A.length>1;if(Ce)for(let Fe=0;Fe<A.length;Fe++)t.bindFramebuffer(r.FRAMEBUFFER,Ye.__webglMultisampledFramebuffer),r.framebufferRenderbuffer(r.FRAMEBUFFER,r.COLOR_ATTACHMENT0+Fe,r.RENDERBUFFER,null),t.bindFramebuffer(r.FRAMEBUFFER,Ye.__webglFramebuffer),r.framebufferTexture2D(r.DRAW_FRAMEBUFFER,r.COLOR_ATTACHMENT0+Fe,r.TEXTURE_2D,null,0);t.bindFramebuffer(r.READ_FRAMEBUFFER,Ye.__webglMultisampledFramebuffer),t.bindFramebuffer(r.DRAW_FRAMEBUFFER,Ye.__webglFramebuffer);for(let Fe=0;Fe<A.length;Fe++){if(N.resolveDepthBuffer&&(N.depthBuffer&&(xe|=r.DEPTH_BUFFER_BIT),N.stencilBuffer&&N.resolveStencilBuffer&&(xe|=r.STENCIL_BUFFER_BIT)),Ce){r.framebufferRenderbuffer(r.READ_FRAMEBUFFER,r.COLOR_ATTACHMENT0,r.RENDERBUFFER,Ye.__webglColorRenderbuffer[Fe]);const pt=s.get(A[Fe]).__webglTexture;r.framebufferTexture2D(r.DRAW_FRAMEBUFFER,r.COLOR_ATTACHMENT0,r.TEXTURE_2D,pt,0)}r.blitFramebuffer(0,0,K,pe,0,0,K,pe,xe,r.NEAREST),h===!0&&(ct.length=0,z.length=0,ct.push(r.COLOR_ATTACHMENT0+Fe),N.depthBuffer&&N.resolveDepthBuffer===!1&&(ct.push(de),z.push(de),r.invalidateFramebuffer(r.DRAW_FRAMEBUFFER,z)),r.invalidateFramebuffer(r.READ_FRAMEBUFFER,ct))}if(t.bindFramebuffer(r.READ_FRAMEBUFFER,null),t.bindFramebuffer(r.DRAW_FRAMEBUFFER,null),Ce)for(let Fe=0;Fe<A.length;Fe++){t.bindFramebuffer(r.FRAMEBUFFER,Ye.__webglMultisampledFramebuffer),r.framebufferRenderbuffer(r.FRAMEBUFFER,r.COLOR_ATTACHMENT0+Fe,r.RENDERBUFFER,Ye.__webglColorRenderbuffer[Fe]);const pt=s.get(A[Fe]).__webglTexture;t.bindFramebuffer(r.FRAMEBUFFER,Ye.__webglFramebuffer),r.framebufferTexture2D(r.DRAW_FRAMEBUFFER,r.COLOR_ATTACHMENT0+Fe,r.TEXTURE_2D,pt,0)}t.bindFramebuffer(r.DRAW_FRAMEBUFFER,Ye.__webglMultisampledFramebuffer)}else if(N.depthBuffer&&N.resolveDepthBuffer===!1&&h){const A=N.stencilBuffer?r.DEPTH_STENCIL_ATTACHMENT:r.DEPTH_ATTACHMENT;r.invalidateFramebuffer(r.DRAW_FRAMEBUFFER,[A])}}}function at(N){return Math.min(a.maxSamples,N.samples)}function ht(N){const A=s.get(N);return N.samples>0&&e.has("WEBGL_multisampled_render_to_texture")===!0&&A.__useRenderToTexture!==!1}function Ze(N){const A=u.render.frame;m.get(N)!==A&&(m.set(N,A),N.update())}function At(N,A){const K=N.colorSpace,pe=N.format,xe=N.type;return N.isCompressedTexture===!0||N.isVideoTexture===!0||K!==br&&K!==Sr&&(Tt.getTransfer(K)===Ut?(pe!==di||xe!==Wi)&&console.warn("THREE.WebGLTextures: sRGB encoded textures have to use RGBAFormat and UnsignedByteType."):console.error("THREE.WebGLTextures: Unsupported texture color space:",K)),A}function Je(N){return typeof HTMLImageElement<"u"&&N instanceof HTMLImageElement?(p.width=N.naturalWidth||N.width,p.height=N.naturalHeight||N.height):typeof VideoFrame<"u"&&N instanceof VideoFrame?(p.width=N.displayWidth,p.height=N.displayHeight):(p.width=N.width,p.height=N.height),p}this.allocateTextureUnit=re,this.resetTextureUnits=C,this.setTexture2D=ae,this.setTexture2DArray=ue,this.setTexture3D=Z,this.setTextureCube=le,this.rebindTextures=Ge,this.setupRenderTarget=dt,this.updateRenderTargetMipmap=gt,this.updateMultisampleRenderTarget=an,this.setupDepthRenderbuffer=Pe,this.setupFrameBufferTexture=fe,this.useMultisampledRTT=ht}function aw(r,e){function t(s,a=Sr){let l;const u=Tt.getTransfer(a);if(s===Wi)return r.UNSIGNED_BYTE;if(s===vd)return r.UNSIGNED_SHORT_4_4_4_4;if(s===xd)return r.UNSIGNED_SHORT_5_5_5_1;if(s===Og)return r.UNSIGNED_INT_5_9_9_9_REV;if(s===Ug)return r.BYTE;if(s===Fg)return r.SHORT;if(s===na)return r.UNSIGNED_SHORT;if(s===_d)return r.INT;if(s===ns)return r.UNSIGNED_INT;if(s===Hi)return r.FLOAT;if(s===sa)return r.HALF_FLOAT;if(s===kg)return r.ALPHA;if(s===Bg)return r.RGB;if(s===di)return r.RGBA;if(s===zg)return r.LUMINANCE;if(s===Hg)return r.LUMINANCE_ALPHA;if(s===Qs)return r.DEPTH_COMPONENT;if(s===ro)return r.DEPTH_STENCIL;if(s===Vg)return r.RED;if(s===yd)return r.RED_INTEGER;if(s===Gg)return r.RG;if(s===Sd)return r.RG_INTEGER;if(s===Md)return r.RGBA_INTEGER;if(s===kl||s===Bl||s===zl||s===Hl)if(u===Ut)if(l=e.get("WEBGL_compressed_texture_s3tc_srgb"),l!==null){if(s===kl)return l.COMPRESSED_SRGB_S3TC_DXT1_EXT;if(s===Bl)return l.COMPRESSED_SRGB_ALPHA_S3TC_DXT1_EXT;if(s===zl)return l.COMPRESSED_SRGB_ALPHA_S3TC_DXT3_EXT;if(s===Hl)return l.COMPRESSED_SRGB_ALPHA_S3TC_DXT5_EXT}else return null;else if(l=e.get("WEBGL_compressed_texture_s3tc"),l!==null){if(s===kl)return l.COMPRESSED_RGB_S3TC_DXT1_EXT;if(s===Bl)return l.COMPRESSED_RGBA_S3TC_DXT1_EXT;if(s===zl)return l.COMPRESSED_RGBA_S3TC_DXT3_EXT;if(s===Hl)return l.COMPRESSED_RGBA_S3TC_DXT5_EXT}else return null;if(s===kf||s===Bf||s===zf||s===Hf)if(l=e.get("WEBGL_compressed_texture_pvrtc"),l!==null){if(s===kf)return l.COMPRESSED_RGB_PVRTC_4BPPV1_IMG;if(s===Bf)return l.COMPRESSED_RGB_PVRTC_2BPPV1_IMG;if(s===zf)return l.COMPRESSED_RGBA_PVRTC_4BPPV1_IMG;if(s===Hf)return l.COMPRESSED_RGBA_PVRTC_2BPPV1_IMG}else return null;if(s===Vf||s===Gf||s===Wf)if(l=e.get("WEBGL_compressed_texture_etc"),l!==null){if(s===Vf||s===Gf)return u===Ut?l.COMPRESSED_SRGB8_ETC2:l.COMPRESSED_RGB8_ETC2;if(s===Wf)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ETC2_EAC:l.COMPRESSED_RGBA8_ETC2_EAC}else return null;if(s===jf||s===Xf||s===Yf||s===qf||s===$f||s===Kf||s===Zf||s===Qf||s===Jf||s===ed||s===td||s===nd||s===id||s===rd)if(l=e.get("WEBGL_compressed_texture_astc"),l!==null){if(s===jf)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ASTC_4x4_KHR:l.COMPRESSED_RGBA_ASTC_4x4_KHR;if(s===Xf)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ASTC_5x4_KHR:l.COMPRESSED_RGBA_ASTC_5x4_KHR;if(s===Yf)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ASTC_5x5_KHR:l.COMPRESSED_RGBA_ASTC_5x5_KHR;if(s===qf)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ASTC_6x5_KHR:l.COMPRESSED_RGBA_ASTC_6x5_KHR;if(s===$f)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ASTC_6x6_KHR:l.COMPRESSED_RGBA_ASTC_6x6_KHR;if(s===Kf)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ASTC_8x5_KHR:l.COMPRESSED_RGBA_ASTC_8x5_KHR;if(s===Zf)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ASTC_8x6_KHR:l.COMPRESSED_RGBA_ASTC_8x6_KHR;if(s===Qf)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ASTC_8x8_KHR:l.COMPRESSED_RGBA_ASTC_8x8_KHR;if(s===Jf)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ASTC_10x5_KHR:l.COMPRESSED_RGBA_ASTC_10x5_KHR;if(s===ed)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ASTC_10x6_KHR:l.COMPRESSED_RGBA_ASTC_10x6_KHR;if(s===td)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ASTC_10x8_KHR:l.COMPRESSED_RGBA_ASTC_10x8_KHR;if(s===nd)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ASTC_10x10_KHR:l.COMPRESSED_RGBA_ASTC_10x10_KHR;if(s===id)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ASTC_12x10_KHR:l.COMPRESSED_RGBA_ASTC_12x10_KHR;if(s===rd)return u===Ut?l.COMPRESSED_SRGB8_ALPHA8_ASTC_12x12_KHR:l.COMPRESSED_RGBA_ASTC_12x12_KHR}else return null;if(s===Vl||s===sd||s===od)if(l=e.get("EXT_texture_compression_bptc"),l!==null){if(s===Vl)return u===Ut?l.COMPRESSED_SRGB_ALPHA_BPTC_UNORM_EXT:l.COMPRESSED_RGBA_BPTC_UNORM_EXT;if(s===sd)return l.COMPRESSED_RGB_BPTC_SIGNED_FLOAT_EXT;if(s===od)return l.COMPRESSED_RGB_BPTC_UNSIGNED_FLOAT_EXT}else return null;if(s===Wg||s===ad||s===ld||s===cd)if(l=e.get("EXT_texture_compression_rgtc"),l!==null){if(s===Vl)return l.COMPRESSED_RED_RGTC1_EXT;if(s===ad)return l.COMPRESSED_SIGNED_RED_RGTC1_EXT;if(s===ld)return l.COMPRESSED_RED_GREEN_RGTC2_EXT;if(s===cd)return l.COMPRESSED_SIGNED_RED_GREEN_RGTC2_EXT}else return null;return s===io?r.UNSIGNED_INT_24_8:r[s]!==void 0?r[s]:null}return{convert:t}}class lw extends Zn{constructor(e=[]){super(),this.isArrayCamera=!0,this.cameras=e}}class Fl extends vn{constructor(){super(),this.isGroup=!0,this.type="Group"}}const cw={type:"move"};class vf{constructor(){this._targetRay=null,this._grip=null,this._hand=null}getHandSpace(){return this._hand===null&&(this._hand=new Fl,this._hand.matrixAutoUpdate=!1,this._hand.visible=!1,this._hand.joints={},this._hand.inputState={pinching:!1}),this._hand}getTargetRaySpace(){return this._targetRay===null&&(this._targetRay=new Fl,this._targetRay.matrixAutoUpdate=!1,this._targetRay.visible=!1,this._targetRay.hasLinearVelocity=!1,this._targetRay.linearVelocity=new Q,this._targetRay.hasAngularVelocity=!1,this._targetRay.angularVelocity=new Q),this._targetRay}getGripSpace(){return this._grip===null&&(this._grip=new Fl,this._grip.matrixAutoUpdate=!1,this._grip.visible=!1,this._grip.hasLinearVelocity=!1,this._grip.linearVelocity=new Q,this._grip.hasAngularVelocity=!1,this._grip.angularVelocity=new Q),this._grip}dispatchEvent(e){return this._targetRay!==null&&this._targetRay.dispatchEvent(e),this._grip!==null&&this._grip.dispatchEvent(e),this._hand!==null&&this._hand.dispatchEvent(e),this}connect(e){if(e&&e.hand){const t=this._hand;if(t)for(const s of e.hand.values())this._getHandJoint(t,s)}return this.dispatchEvent({type:"connected",data:e}),this}disconnect(e){return this.dispatchEvent({type:"disconnected",data:e}),this._targetRay!==null&&(this._targetRay.visible=!1),this._grip!==null&&(this._grip.visible=!1),this._hand!==null&&(this._hand.visible=!1),this}update(e,t,s){let a=null,l=null,u=null;const f=this._targetRay,h=this._grip,p=this._hand;if(e&&t.session.visibilityState!=="visible-blurred"){if(p&&e.hand){u=!0;for(const w of e.hand.values()){const v=t.getJointPose(w,s),y=this._getHandJoint(p,w);v!==null&&(y.matrix.fromArray(v.transform.matrix),y.matrix.decompose(y.position,y.rotation,y.scale),y.matrixWorldNeedsUpdate=!0,y.jointRadius=v.radius),y.visible=v!==null}const m=p.joints["index-finger-tip"],_=p.joints["thumb-tip"],x=m.position.distanceTo(_.position),S=.02,T=.005;p.inputState.pinching&&x>S+T?(p.inputState.pinching=!1,this.dispatchEvent({type:"pinchend",handedness:e.handedness,target:this})):!p.inputState.pinching&&x<=S-T&&(p.inputState.pinching=!0,this.dispatchEvent({type:"pinchstart",handedness:e.handedness,target:this}))}else h!==null&&e.gripSpace&&(l=t.getPose(e.gripSpace,s),l!==null&&(h.matrix.fromArray(l.transform.matrix),h.matrix.decompose(h.position,h.rotation,h.scale),h.matrixWorldNeedsUpdate=!0,l.linearVelocity?(h.hasLinearVelocity=!0,h.linearVelocity.copy(l.linearVelocity)):h.hasLinearVelocity=!1,l.angularVelocity?(h.hasAngularVelocity=!0,h.angularVelocity.copy(l.angularVelocity)):h.hasAngularVelocity=!1));f!==null&&(a=t.getPose(e.targetRaySpace,s),a===null&&l!==null&&(a=l),a!==null&&(f.matrix.fromArray(a.transform.matrix),f.matrix.decompose(f.position,f.rotation,f.scale),f.matrixWorldNeedsUpdate=!0,a.linearVelocity?(f.hasLinearVelocity=!0,f.linearVelocity.copy(a.linearVelocity)):f.hasLinearVelocity=!1,a.angularVelocity?(f.hasAngularVelocity=!0,f.angularVelocity.copy(a.angularVelocity)):f.hasAngularVelocity=!1,this.dispatchEvent(cw)))}return f!==null&&(f.visible=a!==null),h!==null&&(h.visible=l!==null),p!==null&&(p.visible=u!==null),this}_getHandJoint(e,t){if(e.joints[t.jointName]===void 0){const s=new Fl;s.matrixAutoUpdate=!1,s.visible=!1,e.joints[t.jointName]=s,e.add(s)}return e.joints[t.jointName]}}const uw=`
void main() {

	gl_Position = vec4( position, 1.0 );

}`,fw=`
uniform sampler2DArray depthColor;
uniform float depthWidth;
uniform float depthHeight;

void main() {

	vec2 coord = vec2( gl_FragCoord.x / depthWidth, gl_FragCoord.y / depthHeight );

	if ( coord.x >= 1.0 ) {

		gl_FragDepth = texture( depthColor, vec3( coord.x - 1.0, coord.y, 1 ) ).r;

	} else {

		gl_FragDepth = texture( depthColor, vec3( coord.x, coord.y, 0 ) ).r;

	}

}`;class dw{constructor(){this.texture=null,this.mesh=null,this.depthNear=0,this.depthFar=0}init(e,t,s){if(this.texture===null){const a=new Dn,l=e.properties.get(a);l.__webglTexture=t.texture,(t.depthNear!=s.depthNear||t.depthFar!=s.depthFar)&&(this.depthNear=t.depthNear,this.depthFar=t.depthFar),this.texture=a}}getMesh(e){if(this.texture!==null&&this.mesh===null){const t=e.cameras[0].viewport,s=new Cr({vertexShader:uw,fragmentShader:fw,uniforms:{depthColor:{value:this.texture},depthWidth:{value:t.z},depthHeight:{value:t.w}}});this.mesh=new xi(new ic(20,20),s)}return this.mesh}reset(){this.texture=null,this.mesh=null}getDepthTexture(){return this.texture}}class hw extends os{constructor(e,t){super();const s=this;let a=null,l=1,u=null,f="local-floor",h=1,p=null,m=null,_=null,x=null,S=null,T=null;const w=new dw,v=t.getContextAttributes();let y=null,P=null;const b=[],D=[],V=new rt;let O=null;const U=new Zn;U.layers.enable(1),U.viewport=new Vt;const Y=new Zn;Y.layers.enable(2),Y.viewport=new Vt;const ce=[U,Y],E=new lw;E.layers.enable(1),E.layers.enable(2);let C=null,re=null;this.cameraAutoUpdate=!0,this.enabled=!1,this.isPresenting=!1,this.getController=function(J){let fe=b[J];return fe===void 0&&(fe=new vf,b[J]=fe),fe.getTargetRaySpace()},this.getControllerGrip=function(J){let fe=b[J];return fe===void 0&&(fe=new vf,b[J]=fe),fe.getGripSpace()},this.getHand=function(J){let fe=b[J];return fe===void 0&&(fe=new vf,b[J]=fe),fe.getHandSpace()};function ee(J){const fe=D.indexOf(J.inputSource);if(fe===-1)return;const Se=b[fe];Se!==void 0&&(Se.update(J.inputSource,J.frame,p||u),Se.dispatchEvent({type:J.type,data:J.inputSource}))}function ae(){a.removeEventListener("select",ee),a.removeEventListener("selectstart",ee),a.removeEventListener("selectend",ee),a.removeEventListener("squeeze",ee),a.removeEventListener("squeezestart",ee),a.removeEventListener("squeezeend",ee),a.removeEventListener("end",ae),a.removeEventListener("inputsourceschange",ue);for(let J=0;J<b.length;J++){const fe=D[J];fe!==null&&(D[J]=null,b[J].disconnect(fe))}C=null,re=null,w.reset(),e.setRenderTarget(y),S=null,x=null,_=null,a=null,P=null,Ne.stop(),s.isPresenting=!1,e.setPixelRatio(O),e.setSize(V.width,V.height,!1),s.dispatchEvent({type:"sessionend"})}this.setFramebufferScaleFactor=function(J){l=J,s.isPresenting===!0&&console.warn("THREE.WebXRManager: Cannot change framebuffer scale while presenting.")},this.setReferenceSpaceType=function(J){f=J,s.isPresenting===!0&&console.warn("THREE.WebXRManager: Cannot change reference space type while presenting.")},this.getReferenceSpace=function(){return p||u},this.setReferenceSpace=function(J){p=J},this.getBaseLayer=function(){return x!==null?x:S},this.getBinding=function(){return _},this.getFrame=function(){return T},this.getSession=function(){return a},this.setSession=async function(J){if(a=J,a!==null){if(y=e.getRenderTarget(),a.addEventListener("select",ee),a.addEventListener("selectstart",ee),a.addEventListener("selectend",ee),a.addEventListener("squeeze",ee),a.addEventListener("squeezestart",ee),a.addEventListener("squeezeend",ee),a.addEventListener("end",ae),a.addEventListener("inputsourceschange",ue),v.xrCompatible!==!0&&await t.makeXRCompatible(),O=e.getPixelRatio(),e.getSize(V),a.renderState.layers===void 0){const fe={antialias:v.antialias,alpha:!0,depth:v.depth,stencil:v.stencil,framebufferScaleFactor:l};S=new XRWebGLLayer(a,t,fe),a.updateRenderState({baseLayer:S}),e.setPixelRatio(1),e.setSize(S.framebufferWidth,S.framebufferHeight,!1),P=new is(S.framebufferWidth,S.framebufferHeight,{format:di,type:Wi,colorSpace:e.outputColorSpace,stencilBuffer:v.stencil})}else{let fe=null,Se=null,Me=null;v.depth&&(Me=v.stencil?t.DEPTH24_STENCIL8:t.DEPTH_COMPONENT24,fe=v.stencil?ro:Qs,Se=v.stencil?io:ns);const Pe={colorFormat:t.RGBA8,depthFormat:Me,scaleFactor:l};_=new XRWebGLBinding(a,t),x=_.createProjectionLayer(Pe),a.updateRenderState({layers:[x]}),e.setPixelRatio(1),e.setSize(x.textureWidth,x.textureHeight,!1),P=new is(x.textureWidth,x.textureHeight,{format:di,type:Wi,depthTexture:new a_(x.textureWidth,x.textureHeight,Se,void 0,void 0,void 0,void 0,void 0,void 0,fe),stencilBuffer:v.stencil,colorSpace:e.outputColorSpace,samples:v.antialias?4:0,resolveDepthBuffer:x.ignoreDepthValues===!1})}P.isXRRenderTarget=!0,this.setFoveation(h),p=null,u=await a.requestReferenceSpace(f),Ne.setContext(a),Ne.start(),s.isPresenting=!0,s.dispatchEvent({type:"sessionstart"})}},this.getEnvironmentBlendMode=function(){if(a!==null)return a.environmentBlendMode},this.getDepthTexture=function(){return w.getDepthTexture()};function ue(J){for(let fe=0;fe<J.removed.length;fe++){const Se=J.removed[fe],Me=D.indexOf(Se);Me>=0&&(D[Me]=null,b[Me].disconnect(Se))}for(let fe=0;fe<J.added.length;fe++){const Se=J.added[fe];let Me=D.indexOf(Se);if(Me===-1){for(let Ge=0;Ge<b.length;Ge++)if(Ge>=D.length){D.push(Se),Me=Ge;break}else if(D[Ge]===null){D[Ge]=Se,Me=Ge;break}if(Me===-1)break}const Pe=b[Me];Pe&&Pe.connect(Se)}}const Z=new Q,le=new Q;function F(J,fe,Se){Z.setFromMatrixPosition(fe.matrixWorld),le.setFromMatrixPosition(Se.matrixWorld);const Me=Z.distanceTo(le),Pe=fe.projectionMatrix.elements,Ge=Se.projectionMatrix.elements,dt=Pe[14]/(Pe[10]-1),gt=Pe[14]/(Pe[10]+1),ct=(Pe[9]+1)/Pe[5],z=(Pe[9]-1)/Pe[5],an=(Pe[8]-1)/Pe[0],at=(Ge[8]+1)/Ge[0],ht=dt*an,Ze=dt*at,At=Me/(-an+at),Je=At*-an;if(fe.matrixWorld.decompose(J.position,J.quaternion,J.scale),J.translateX(Je),J.translateZ(At),J.matrixWorld.compose(J.position,J.quaternion,J.scale),J.matrixWorldInverse.copy(J.matrixWorld).invert(),Pe[10]===-1)J.projectionMatrix.copy(fe.projectionMatrix),J.projectionMatrixInverse.copy(fe.projectionMatrixInverse);else{const N=dt+At,A=gt+At,K=ht-Je,pe=Ze+(Me-Je),xe=ct*gt/A*N,de=z*gt/A*N;J.projectionMatrix.makePerspective(K,pe,xe,de,N,A),J.projectionMatrixInverse.copy(J.projectionMatrix).invert()}}function se(J,fe){fe===null?J.matrixWorld.copy(J.matrix):J.matrixWorld.multiplyMatrices(fe.matrixWorld,J.matrix),J.matrixWorldInverse.copy(J.matrixWorld).invert()}this.updateCamera=function(J){if(a===null)return;let fe=J.near,Se=J.far;w.texture!==null&&(w.depthNear>0&&(fe=w.depthNear),w.depthFar>0&&(Se=w.depthFar)),E.near=Y.near=U.near=fe,E.far=Y.far=U.far=Se,(C!==E.near||re!==E.far)&&(a.updateRenderState({depthNear:E.near,depthFar:E.far}),C=E.near,re=E.far);const Me=J.parent,Pe=E.cameras;se(E,Me);for(let Ge=0;Ge<Pe.length;Ge++)se(Pe[Ge],Me);Pe.length===2?F(E,U,Y):E.projectionMatrix.copy(U.projectionMatrix),L(J,E,Me)};function L(J,fe,Se){Se===null?J.matrix.copy(fe.matrixWorld):(J.matrix.copy(Se.matrixWorld),J.matrix.invert(),J.matrix.multiply(fe.matrixWorld)),J.matrix.decompose(J.position,J.quaternion,J.scale),J.updateMatrixWorld(!0),J.projectionMatrix.copy(fe.projectionMatrix),J.projectionMatrixInverse.copy(fe.projectionMatrixInverse),J.isPerspectiveCamera&&(J.fov=ud*2*Math.atan(1/J.projectionMatrix.elements[5]),J.zoom=1)}this.getCamera=function(){return E},this.getFoveation=function(){if(!(x===null&&S===null))return h},this.setFoveation=function(J){h=J,x!==null&&(x.fixedFoveation=J),S!==null&&S.fixedFoveation!==void 0&&(S.fixedFoveation=J)},this.hasDepthSensing=function(){return w.texture!==null},this.getDepthSensingMesh=function(){return w.getMesh(E)};let X=null;function ve(J,fe){if(m=fe.getViewerPose(p||u),T=fe,m!==null){const Se=m.views;S!==null&&(e.setRenderTargetFramebuffer(P,S.framebuffer),e.setRenderTarget(P));let Me=!1;Se.length!==E.cameras.length&&(E.cameras.length=0,Me=!0);for(let Ge=0;Ge<Se.length;Ge++){const dt=Se[Ge];let gt=null;if(S!==null)gt=S.getViewport(dt);else{const z=_.getViewSubImage(x,dt);gt=z.viewport,Ge===0&&(e.setRenderTargetTextures(P,z.colorTexture,x.ignoreDepthValues?void 0:z.depthStencilTexture),e.setRenderTarget(P))}let ct=ce[Ge];ct===void 0&&(ct=new Zn,ct.layers.enable(Ge),ct.viewport=new Vt,ce[Ge]=ct),ct.matrix.fromArray(dt.transform.matrix),ct.matrix.decompose(ct.position,ct.quaternion,ct.scale),ct.projectionMatrix.fromArray(dt.projectionMatrix),ct.projectionMatrixInverse.copy(ct.projectionMatrix).invert(),ct.viewport.set(gt.x,gt.y,gt.width,gt.height),Ge===0&&(E.matrix.copy(ct.matrix),E.matrix.decompose(E.position,E.quaternion,E.scale)),Me===!0&&E.cameras.push(ct)}const Pe=a.enabledFeatures;if(Pe&&Pe.includes("depth-sensing")){const Ge=_.getDepthInformation(Se[0]);Ge&&Ge.isValid&&Ge.texture&&w.init(e,Ge,a.renderState)}}for(let Se=0;Se<b.length;Se++){const Me=D[Se],Pe=b[Se];Me!==null&&Pe!==void 0&&Pe.update(Me,fe,p||u)}X&&X(J,fe),fe.detectedPlanes&&s.dispatchEvent({type:"planesdetected",data:fe}),T=null}const Ne=new s_;Ne.setAnimationLoop(ve),this.setAnimationLoop=function(J){X=J},this.dispose=function(){}}}const $r=new Si,pw=new Gt;function mw(r,e){function t(v,y){v.matrixAutoUpdate===!0&&v.updateMatrix(),y.value.copy(v.matrix)}function s(v,y){y.color.getRGB(v.fogColor.value,n_(r)),y.isFog?(v.fogNear.value=y.near,v.fogFar.value=y.far):y.isFogExp2&&(v.fogDensity.value=y.density)}function a(v,y,P,b,D){y.isMeshBasicMaterial||y.isMeshLambertMaterial?l(v,y):y.isMeshToonMaterial?(l(v,y),_(v,y)):y.isMeshPhongMaterial?(l(v,y),m(v,y)):y.isMeshStandardMaterial?(l(v,y),x(v,y),y.isMeshPhysicalMaterial&&S(v,y,D)):y.isMeshMatcapMaterial?(l(v,y),T(v,y)):y.isMeshDepthMaterial?l(v,y):y.isMeshDistanceMaterial?(l(v,y),w(v,y)):y.isMeshNormalMaterial?l(v,y):y.isLineBasicMaterial?(u(v,y),y.isLineDashedMaterial&&f(v,y)):y.isPointsMaterial?h(v,y,P,b):y.isSpriteMaterial?p(v,y):y.isShadowMaterial?(v.color.value.copy(y.color),v.opacity.value=y.opacity):y.isShaderMaterial&&(y.uniformsNeedUpdate=!1)}function l(v,y){v.opacity.value=y.opacity,y.color&&v.diffuse.value.copy(y.color),y.emissive&&v.emissive.value.copy(y.emissive).multiplyScalar(y.emissiveIntensity),y.map&&(v.map.value=y.map,t(y.map,v.mapTransform)),y.alphaMap&&(v.alphaMap.value=y.alphaMap,t(y.alphaMap,v.alphaMapTransform)),y.bumpMap&&(v.bumpMap.value=y.bumpMap,t(y.bumpMap,v.bumpMapTransform),v.bumpScale.value=y.bumpScale,y.side===Ln&&(v.bumpScale.value*=-1)),y.normalMap&&(v.normalMap.value=y.normalMap,t(y.normalMap,v.normalMapTransform),v.normalScale.value.copy(y.normalScale),y.side===Ln&&v.normalScale.value.negate()),y.displacementMap&&(v.displacementMap.value=y.displacementMap,t(y.displacementMap,v.displacementMapTransform),v.displacementScale.value=y.displacementScale,v.displacementBias.value=y.displacementBias),y.emissiveMap&&(v.emissiveMap.value=y.emissiveMap,t(y.emissiveMap,v.emissiveMapTransform)),y.specularMap&&(v.specularMap.value=y.specularMap,t(y.specularMap,v.specularMapTransform)),y.alphaTest>0&&(v.alphaTest.value=y.alphaTest);const P=e.get(y),b=P.envMap,D=P.envMapRotation;b&&(v.envMap.value=b,$r.copy(D),$r.x*=-1,$r.y*=-1,$r.z*=-1,b.isCubeTexture&&b.isRenderTargetTexture===!1&&($r.y*=-1,$r.z*=-1),v.envMapRotation.value.setFromMatrix4(pw.makeRotationFromEuler($r)),v.flipEnvMap.value=b.isCubeTexture&&b.isRenderTargetTexture===!1?-1:1,v.reflectivity.value=y.reflectivity,v.ior.value=y.ior,v.refractionRatio.value=y.refractionRatio),y.lightMap&&(v.lightMap.value=y.lightMap,v.lightMapIntensity.value=y.lightMapIntensity,t(y.lightMap,v.lightMapTransform)),y.aoMap&&(v.aoMap.value=y.aoMap,v.aoMapIntensity.value=y.aoMapIntensity,t(y.aoMap,v.aoMapTransform))}function u(v,y){v.diffuse.value.copy(y.color),v.opacity.value=y.opacity,y.map&&(v.map.value=y.map,t(y.map,v.mapTransform))}function f(v,y){v.dashSize.value=y.dashSize,v.totalSize.value=y.dashSize+y.gapSize,v.scale.value=y.scale}function h(v,y,P,b){v.diffuse.value.copy(y.color),v.opacity.value=y.opacity,v.size.value=y.size*P,v.scale.value=b*.5,y.map&&(v.map.value=y.map,t(y.map,v.uvTransform)),y.alphaMap&&(v.alphaMap.value=y.alphaMap,t(y.alphaMap,v.alphaMapTransform)),y.alphaTest>0&&(v.alphaTest.value=y.alphaTest)}function p(v,y){v.diffuse.value.copy(y.color),v.opacity.value=y.opacity,v.rotation.value=y.rotation,y.map&&(v.map.value=y.map,t(y.map,v.mapTransform)),y.alphaMap&&(v.alphaMap.value=y.alphaMap,t(y.alphaMap,v.alphaMapTransform)),y.alphaTest>0&&(v.alphaTest.value=y.alphaTest)}function m(v,y){v.specular.value.copy(y.specular),v.shininess.value=Math.max(y.shininess,1e-4)}function _(v,y){y.gradientMap&&(v.gradientMap.value=y.gradientMap)}function x(v,y){v.metalness.value=y.metalness,y.metalnessMap&&(v.metalnessMap.value=y.metalnessMap,t(y.metalnessMap,v.metalnessMapTransform)),v.roughness.value=y.roughness,y.roughnessMap&&(v.roughnessMap.value=y.roughnessMap,t(y.roughnessMap,v.roughnessMapTransform)),y.envMap&&(v.envMapIntensity.value=y.envMapIntensity)}function S(v,y,P){v.ior.value=y.ior,y.sheen>0&&(v.sheenColor.value.copy(y.sheenColor).multiplyScalar(y.sheen),v.sheenRoughness.value=y.sheenRoughness,y.sheenColorMap&&(v.sheenColorMap.value=y.sheenColorMap,t(y.sheenColorMap,v.sheenColorMapTransform)),y.sheenRoughnessMap&&(v.sheenRoughnessMap.value=y.sheenRoughnessMap,t(y.sheenRoughnessMap,v.sheenRoughnessMapTransform))),y.clearcoat>0&&(v.clearcoat.value=y.clearcoat,v.clearcoatRoughness.value=y.clearcoatRoughness,y.clearcoatMap&&(v.clearcoatMap.value=y.clearcoatMap,t(y.clearcoatMap,v.clearcoatMapTransform)),y.clearcoatRoughnessMap&&(v.clearcoatRoughnessMap.value=y.clearcoatRoughnessMap,t(y.clearcoatRoughnessMap,v.clearcoatRoughnessMapTransform)),y.clearcoatNormalMap&&(v.clearcoatNormalMap.value=y.clearcoatNormalMap,t(y.clearcoatNormalMap,v.clearcoatNormalMapTransform),v.clearcoatNormalScale.value.copy(y.clearcoatNormalScale),y.side===Ln&&v.clearcoatNormalScale.value.negate())),y.dispersion>0&&(v.dispersion.value=y.dispersion),y.iridescence>0&&(v.iridescence.value=y.iridescence,v.iridescenceIOR.value=y.iridescenceIOR,v.iridescenceThicknessMinimum.value=y.iridescenceThicknessRange[0],v.iridescenceThicknessMaximum.value=y.iridescenceThicknessRange[1],y.iridescenceMap&&(v.iridescenceMap.value=y.iridescenceMap,t(y.iridescenceMap,v.iridescenceMapTransform)),y.iridescenceThicknessMap&&(v.iridescenceThicknessMap.value=y.iridescenceThicknessMap,t(y.iridescenceThicknessMap,v.iridescenceThicknessMapTransform))),y.transmission>0&&(v.transmission.value=y.transmission,v.transmissionSamplerMap.value=P.texture,v.transmissionSamplerSize.value.set(P.width,P.height),y.transmissionMap&&(v.transmissionMap.value=y.transmissionMap,t(y.transmissionMap,v.transmissionMapTransform)),v.thickness.value=y.thickness,y.thicknessMap&&(v.thicknessMap.value=y.thicknessMap,t(y.thicknessMap,v.thicknessMapTransform)),v.attenuationDistance.value=y.attenuationDistance,v.attenuationColor.value.copy(y.attenuationColor)),y.anisotropy>0&&(v.anisotropyVector.value.set(y.anisotropy*Math.cos(y.anisotropyRotation),y.anisotropy*Math.sin(y.anisotropyRotation)),y.anisotropyMap&&(v.anisotropyMap.value=y.anisotropyMap,t(y.anisotropyMap,v.anisotropyMapTransform))),v.specularIntensity.value=y.specularIntensity,v.specularColor.value.copy(y.specularColor),y.specularColorMap&&(v.specularColorMap.value=y.specularColorMap,t(y.specularColorMap,v.specularColorMapTransform)),y.specularIntensityMap&&(v.specularIntensityMap.value=y.specularIntensityMap,t(y.specularIntensityMap,v.specularIntensityMapTransform))}function T(v,y){y.matcap&&(v.matcap.value=y.matcap)}function w(v,y){const P=e.get(y).light;v.referencePosition.value.setFromMatrixPosition(P.matrixWorld),v.nearDistance.value=P.shadow.camera.near,v.farDistance.value=P.shadow.camera.far}return{refreshFogUniforms:s,refreshMaterialUniforms:a}}function gw(r,e,t,s){let a={},l={},u=[];const f=r.getParameter(r.MAX_UNIFORM_BUFFER_BINDINGS);function h(P,b){const D=b.program;s.uniformBlockBinding(P,D)}function p(P,b){let D=a[P.id];D===void 0&&(T(P),D=m(P),a[P.id]=D,P.addEventListener("dispose",v));const V=b.program;s.updateUBOMapping(P,V);const O=e.render.frame;l[P.id]!==O&&(x(P),l[P.id]=O)}function m(P){const b=_();P.__bindingPointIndex=b;const D=r.createBuffer(),V=P.__size,O=P.usage;return r.bindBuffer(r.UNIFORM_BUFFER,D),r.bufferData(r.UNIFORM_BUFFER,V,O),r.bindBuffer(r.UNIFORM_BUFFER,null),r.bindBufferBase(r.UNIFORM_BUFFER,b,D),D}function _(){for(let P=0;P<f;P++)if(u.indexOf(P)===-1)return u.push(P),P;return console.error("THREE.WebGLRenderer: Maximum number of simultaneously usable uniforms groups reached."),0}function x(P){const b=a[P.id],D=P.uniforms,V=P.__cache;r.bindBuffer(r.UNIFORM_BUFFER,b);for(let O=0,U=D.length;O<U;O++){const Y=Array.isArray(D[O])?D[O]:[D[O]];for(let ce=0,E=Y.length;ce<E;ce++){const C=Y[ce];if(S(C,O,ce,V)===!0){const re=C.__offset,ee=Array.isArray(C.value)?C.value:[C.value];let ae=0;for(let ue=0;ue<ee.length;ue++){const Z=ee[ue],le=w(Z);typeof Z=="number"||typeof Z=="boolean"?(C.__data[0]=Z,r.bufferSubData(r.UNIFORM_BUFFER,re+ae,C.__data)):Z.isMatrix3?(C.__data[0]=Z.elements[0],C.__data[1]=Z.elements[1],C.__data[2]=Z.elements[2],C.__data[3]=0,C.__data[4]=Z.elements[3],C.__data[5]=Z.elements[4],C.__data[6]=Z.elements[5],C.__data[7]=0,C.__data[8]=Z.elements[6],C.__data[9]=Z.elements[7],C.__data[10]=Z.elements[8],C.__data[11]=0):(Z.toArray(C.__data,ae),ae+=le.storage/Float32Array.BYTES_PER_ELEMENT)}r.bufferSubData(r.UNIFORM_BUFFER,re,C.__data)}}}r.bindBuffer(r.UNIFORM_BUFFER,null)}function S(P,b,D,V){const O=P.value,U=b+"_"+D;if(V[U]===void 0)return typeof O=="number"||typeof O=="boolean"?V[U]=O:V[U]=O.clone(),!0;{const Y=V[U];if(typeof O=="number"||typeof O=="boolean"){if(Y!==O)return V[U]=O,!0}else if(Y.equals(O)===!1)return Y.copy(O),!0}return!1}function T(P){const b=P.uniforms;let D=0;const V=16;for(let U=0,Y=b.length;U<Y;U++){const ce=Array.isArray(b[U])?b[U]:[b[U]];for(let E=0,C=ce.length;E<C;E++){const re=ce[E],ee=Array.isArray(re.value)?re.value:[re.value];for(let ae=0,ue=ee.length;ae<ue;ae++){const Z=ee[ae],le=w(Z),F=D%V,se=F%le.boundary,L=F+se;D+=se,L!==0&&V-L<le.storage&&(D+=V-L),re.__data=new Float32Array(le.storage/Float32Array.BYTES_PER_ELEMENT),re.__offset=D,D+=le.storage}}}const O=D%V;return O>0&&(D+=V-O),P.__size=D,P.__cache={},this}function w(P){const b={boundary:0,storage:0};return typeof P=="number"||typeof P=="boolean"?(b.boundary=4,b.storage=4):P.isVector2?(b.boundary=8,b.storage=8):P.isVector3||P.isColor?(b.boundary=16,b.storage=12):P.isVector4?(b.boundary=16,b.storage=16):P.isMatrix3?(b.boundary=48,b.storage=48):P.isMatrix4?(b.boundary=64,b.storage=64):P.isTexture?console.warn("THREE.WebGLRenderer: Texture samplers can not be part of an uniforms group."):console.warn("THREE.WebGLRenderer: Unsupported uniform value type.",P),b}function v(P){const b=P.target;b.removeEventListener("dispose",v);const D=u.indexOf(b.__bindingPointIndex);u.splice(D,1),r.deleteBuffer(a[b.id]),delete a[b.id],delete l[b.id]}function y(){for(const P in a)r.deleteBuffer(a[P]);u=[],a={},l={}}return{bind:h,update:p,dispose:y}}class _w{constructor(e={}){const{canvas:t=ry(),context:s=null,depth:a=!0,stencil:l=!1,alpha:u=!1,antialias:f=!1,premultipliedAlpha:h=!0,preserveDrawingBuffer:p=!1,powerPreference:m="default",failIfMajorPerformanceCaveat:_=!1}=e;this.isWebGLRenderer=!0;let x;if(s!==null){if(typeof WebGLRenderingContext<"u"&&s instanceof WebGLRenderingContext)throw new Error("THREE.WebGLRenderer: WebGL 1 is not supported since r163.");x=s.getContextAttributes().alpha}else x=u;const S=new Uint32Array(4),T=new Int32Array(4);let w=null,v=null;const y=[],P=[];this.domElement=t,this.debug={checkShaderErrors:!0,onShaderError:null},this.autoClear=!0,this.autoClearColor=!0,this.autoClearDepth=!0,this.autoClearStencil=!0,this.sortObjects=!0,this.clippingPlanes=[],this.localClippingEnabled=!1,this._outputColorSpace=ci,this.toneMapping=wr,this.toneMappingExposure=1;const b=this;let D=!1,V=0,O=0,U=null,Y=-1,ce=null;const E=new Vt,C=new Vt;let re=null;const ee=new vt(0);let ae=0,ue=t.width,Z=t.height,le=1,F=null,se=null;const L=new Vt(0,0,ue,Z),X=new Vt(0,0,ue,Z);let ve=!1;const Ne=new wd;let J=!1,fe=!1;const Se=new Gt,Me=new Gt,Pe=new Q,Ge=new Vt,dt={background:null,fog:null,environment:null,overrideMaterial:null,isScene:!0};let gt=!1;function ct(){return U===null?le:1}let z=s;function an(R,W){return t.getContext(R,W)}try{const R={alpha:!0,depth:a,stencil:l,antialias:f,premultipliedAlpha:h,preserveDrawingBuffer:p,powerPreference:m,failIfMajorPerformanceCaveat:_};if("setAttribute"in t&&t.setAttribute("data-engine",`three.js r${gd}`),t.addEventListener("webglcontextlost",me,!1),t.addEventListener("webglcontextrestored",Re,!1),t.addEventListener("webglcontextcreationerror",Ie,!1),z===null){const W="webgl2";if(z=an(W,R),z===null)throw an(W)?new Error("Error creating WebGL context with your selected attributes."):new Error("Error creating WebGL context.")}}catch(R){throw console.error("THREE.WebGLRenderer: "+R.message),R}let at,ht,Ze,At,Je,N,A,K,pe,xe,de,Ye,Ce,Fe,pt,Ee,Oe,tt,et,Be,ut,it,Et,G;function Le(){at=new ME(z),at.init(),it=new aw(z,at),ht=new gE(z,at,e,it),Ze=new rw(z),ht.reverseDepthBuffer&&Ze.buffers.depth.setReversed(!0),At=new wE(z),Je=new GT,N=new ow(z,at,Ze,Je,ht,it,At),A=new vE(b),K=new SE(b),pe=new Dy(z),Et=new pE(z,pe),xe=new EE(z,pe,At,Et),de=new CE(z,xe,pe,At),et=new AE(z,ht,N),Ee=new _E(Je),Ye=new VT(b,A,K,at,ht,Et,Ee),Ce=new mw(b,Je),Fe=new jT,pt=new ZT(at),tt=new hE(b,A,K,Ze,de,x,h),Oe=new nw(b,de,ht),G=new gw(z,At,ht,Ze),Be=new mE(z,at,At),ut=new TE(z,at,At),At.programs=Ye.programs,b.capabilities=ht,b.extensions=at,b.properties=Je,b.renderLists=Fe,b.shadowMap=Oe,b.state=Ze,b.info=At}Le();const oe=new hw(b,z);this.xr=oe,this.getContext=function(){return z},this.getContextAttributes=function(){return z.getContextAttributes()},this.forceContextLoss=function(){const R=at.get("WEBGL_lose_context");R&&R.loseContext()},this.forceContextRestore=function(){const R=at.get("WEBGL_lose_context");R&&R.restoreContext()},this.getPixelRatio=function(){return le},this.setPixelRatio=function(R){R!==void 0&&(le=R,this.setSize(ue,Z,!1))},this.getSize=function(R){return R.set(ue,Z)},this.setSize=function(R,W,te=!0){if(oe.isPresenting){console.warn("THREE.WebGLRenderer: Can't change size while VR device is presenting.");return}ue=R,Z=W,t.width=Math.floor(R*le),t.height=Math.floor(W*le),te===!0&&(t.style.width=R+"px",t.style.height=W+"px"),this.setViewport(0,0,R,W)},this.getDrawingBufferSize=function(R){return R.set(ue*le,Z*le).floor()},this.setDrawingBufferSize=function(R,W,te){ue=R,Z=W,le=te,t.width=Math.floor(R*te),t.height=Math.floor(W*te),this.setViewport(0,0,R,W)},this.getCurrentViewport=function(R){return R.copy(E)},this.getViewport=function(R){return R.copy(L)},this.setViewport=function(R,W,te,ne){R.isVector4?L.set(R.x,R.y,R.z,R.w):L.set(R,W,te,ne),Ze.viewport(E.copy(L).multiplyScalar(le).round())},this.getScissor=function(R){return R.copy(X)},this.setScissor=function(R,W,te,ne){R.isVector4?X.set(R.x,R.y,R.z,R.w):X.set(R,W,te,ne),Ze.scissor(C.copy(X).multiplyScalar(le).round())},this.getScissorTest=function(){return ve},this.setScissorTest=function(R){Ze.setScissorTest(ve=R)},this.setOpaqueSort=function(R){F=R},this.setTransparentSort=function(R){se=R},this.getClearColor=function(R){return R.copy(tt.getClearColor())},this.setClearColor=function(){tt.setClearColor.apply(tt,arguments)},this.getClearAlpha=function(){return tt.getClearAlpha()},this.setClearAlpha=function(){tt.setClearAlpha.apply(tt,arguments)},this.clear=function(R=!0,W=!0,te=!0){let ne=0;if(R){let j=!1;if(U!==null){const we=U.texture.format;j=we===Md||we===Sd||we===yd}if(j){const we=U.texture.type,De=we===Wi||we===ns||we===na||we===io||we===vd||we===xd,Ae=tt.getClearColor(),We=tt.getClearAlpha(),Ke=Ae.r,Qe=Ae.g,je=Ae.b;De?(S[0]=Ke,S[1]=Qe,S[2]=je,S[3]=We,z.clearBufferuiv(z.COLOR,0,S)):(T[0]=Ke,T[1]=Qe,T[2]=je,T[3]=We,z.clearBufferiv(z.COLOR,0,T))}else ne|=z.COLOR_BUFFER_BIT}W&&(ne|=z.DEPTH_BUFFER_BIT,z.clearDepth(this.capabilities.reverseDepthBuffer?0:1)),te&&(ne|=z.STENCIL_BUFFER_BIT,this.state.buffers.stencil.setMask(4294967295)),z.clear(ne)},this.clearColor=function(){this.clear(!0,!1,!1)},this.clearDepth=function(){this.clear(!1,!0,!1)},this.clearStencil=function(){this.clear(!1,!1,!0)},this.dispose=function(){t.removeEventListener("webglcontextlost",me,!1),t.removeEventListener("webglcontextrestored",Re,!1),t.removeEventListener("webglcontextcreationerror",Ie,!1),Fe.dispose(),pt.dispose(),Je.dispose(),A.dispose(),K.dispose(),de.dispose(),Et.dispose(),G.dispose(),Ye.dispose(),oe.dispose(),oe.removeEventListener("sessionstart",Xi),oe.removeEventListener("sessionend",as),Nn.stop()};function me(R){R.preventDefault(),console.log("THREE.WebGLRenderer: Context Lost."),D=!0}function Re(){console.log("THREE.WebGLRenderer: Context Restored."),D=!1;const R=At.autoReset,W=Oe.enabled,te=Oe.autoUpdate,ne=Oe.needsUpdate,j=Oe.type;Le(),At.autoReset=R,Oe.enabled=W,Oe.autoUpdate=te,Oe.needsUpdate=ne,Oe.type=j}function Ie(R){console.error("THREE.WebGLRenderer: A WebGL context could not be created. Reason: ",R.statusMessage)}function ft(R){const W=R.target;W.removeEventListener("dispose",ft),kt(W)}function kt(R){ln(R),Je.remove(R)}function ln(R){const W=Je.get(R).programs;W!==void 0&&(W.forEach(function(te){Ye.releaseProgram(te)}),R.isShaderMaterial&&Ye.releaseShaderCache(R))}this.renderBufferDirect=function(R,W,te,ne,j,we){W===null&&(W=dt);const De=j.isMesh&&j.matrixWorld.determinant()<0,Ae=Ei(R,W,te,ne,j);Ze.setMaterial(ne,De);let We=te.index,Ke=1;if(ne.wireframe===!0){if(We=xe.getWireframeAttribute(te),We===void 0)return;Ke=2}const Qe=te.drawRange,je=te.attributes.position;let Mt=Qe.start*Ke,Ct=(Qe.start+Qe.count)*Ke;we!==null&&(Mt=Math.max(Mt,we.start*Ke),Ct=Math.min(Ct,(we.start+we.count)*Ke)),We!==null?(Mt=Math.max(Mt,0),Ct=Math.min(Ct,We.count)):je!=null&&(Mt=Math.max(Mt,0),Ct=Math.min(Ct,je.count));const bt=Ct-Mt;if(bt<0||bt===1/0)return;Et.setup(j,ne,Ae,te,We);let Ft,xt=Be;if(We!==null&&(Ft=pe.get(We),xt=ut,xt.setIndex(Ft)),j.isMesh)ne.wireframe===!0?(Ze.setLineWidth(ne.wireframeLinewidth*ct()),xt.setMode(z.LINES)):xt.setMode(z.TRIANGLES);else if(j.isLine){let ke=ne.linewidth;ke===void 0&&(ke=1),Ze.setLineWidth(ke*ct()),j.isLineSegments?xt.setMode(z.LINES):j.isLineLoop?xt.setMode(z.LINE_LOOP):xt.setMode(z.LINE_STRIP)}else j.isPoints?xt.setMode(z.POINTS):j.isSprite&&xt.setMode(z.TRIANGLES);if(j.isBatchedMesh)if(j._multiDrawInstances!==null)xt.renderMultiDrawInstances(j._multiDrawStarts,j._multiDrawCounts,j._multiDrawCount,j._multiDrawInstances);else if(at.get("WEBGL_multi_draw"))xt.renderMultiDraw(j._multiDrawStarts,j._multiDrawCounts,j._multiDrawCount);else{const ke=j._multiDrawStarts,qt=j._multiDrawCounts,yt=j._multiDrawCount,In=We?pe.get(We).bytesPerElement:1,Jn=Je.get(ne).currentProgram.getUniforms();for(let tn=0;tn<yt;tn++)Jn.setValue(z,"_gl_DrawID",tn),xt.render(ke[tn]/In,qt[tn])}else if(j.isInstancedMesh)xt.renderInstances(Mt,bt,j.count);else if(te.isInstancedBufferGeometry){const ke=te._maxInstanceCount!==void 0?te._maxInstanceCount:1/0,qt=Math.min(te.instanceCount,ke);xt.renderInstances(Mt,bt,qt)}else xt.render(Mt,bt)};function mt(R,W,te){R.transparent===!0&&R.side===zi&&R.forceSinglePass===!1?(R.side=Ln,R.needsUpdate=!0,cs(R,W,te),R.side=Ar,R.needsUpdate=!0,cs(R,W,te),R.side=zi):cs(R,W,te)}this.compile=function(R,W,te=null){te===null&&(te=R),v=pt.get(te),v.init(W),P.push(v),te.traverseVisible(function(j){j.isLight&&j.layers.test(W.layers)&&(v.pushLight(j),j.castShadow&&v.pushShadow(j))}),R!==te&&R.traverseVisible(function(j){j.isLight&&j.layers.test(W.layers)&&(v.pushLight(j),j.castShadow&&v.pushShadow(j))}),v.setupLights();const ne=new Set;return R.traverse(function(j){if(!(j.isMesh||j.isPoints||j.isLine||j.isSprite))return;const we=j.material;if(we)if(Array.isArray(we))for(let De=0;De<we.length;De++){const Ae=we[De];mt(Ae,te,j),ne.add(Ae)}else mt(we,te,j),ne.add(we)}),P.pop(),v=null,ne},this.compileAsync=function(R,W,te=null){const ne=this.compile(R,W,te);return new Promise(j=>{function we(){if(ne.forEach(function(De){Je.get(De).currentProgram.isReady()&&ne.delete(De)}),ne.size===0){j(R);return}setTimeout(we,10)}at.get("KHR_parallel_shader_compile")!==null?we():setTimeout(we,10)})};let en=null;function Gn(R){en&&en(R)}function Xi(){Nn.stop()}function as(){Nn.start()}const Nn=new s_;Nn.setAnimationLoop(Gn),typeof self<"u"&&Nn.setContext(self),this.setAnimationLoop=function(R){en=R,oe.setAnimationLoop(R),R===null?Nn.stop():Nn.start()},oe.addEventListener("sessionstart",Xi),oe.addEventListener("sessionend",as),this.render=function(R,W){if(W!==void 0&&W.isCamera!==!0){console.error("THREE.WebGLRenderer.render: camera is not an instance of THREE.Camera.");return}if(D===!0)return;if(R.matrixWorldAutoUpdate===!0&&R.updateMatrixWorld(),W.parent===null&&W.matrixWorldAutoUpdate===!0&&W.updateMatrixWorld(),oe.enabled===!0&&oe.isPresenting===!0&&(oe.cameraAutoUpdate===!0&&oe.updateCamera(W),W=oe.getCamera()),R.isScene===!0&&R.onBeforeRender(b,R,W,U),v=pt.get(R,P.length),v.init(W),P.push(v),Me.multiplyMatrices(W.projectionMatrix,W.matrixWorldInverse),Ne.setFromProjectionMatrix(Me),fe=this.localClippingEnabled,J=Ee.init(this.clippingPlanes,fe),w=Fe.get(R,y.length),w.init(),y.push(w),oe.enabled===!0&&oe.isPresenting===!0){const we=b.xr.getDepthSensingMesh();we!==null&&co(we,W,-1/0,b.sortObjects)}co(R,W,0,b.sortObjects),w.finish(),b.sortObjects===!0&&w.sort(F,se),gt=oe.enabled===!1||oe.isPresenting===!1||oe.hasDepthSensing()===!1,gt&&tt.addToRenderList(w,R),this.info.render.frame++,J===!0&&Ee.beginShadows();const te=v.state.shadowsArray;Oe.render(te,R,W),J===!0&&Ee.endShadows(),this.info.autoReset===!0&&this.info.reset();const ne=w.opaque,j=w.transmissive;if(v.setupLights(),W.isArrayCamera){const we=W.cameras;if(j.length>0)for(let De=0,Ae=we.length;De<Ae;De++){const We=we[De];Pr(ne,j,R,We)}gt&&tt.render(R);for(let De=0,Ae=we.length;De<Ae;De++){const We=we[De];Yi(w,R,We,We.viewport)}}else j.length>0&&Pr(ne,j,R,W),gt&&tt.render(R),Yi(w,R,W);U!==null&&(N.updateMultisampleRenderTarget(U),N.updateRenderTargetMipmap(U)),R.isScene===!0&&R.onAfterRender(b,R,W),Et.resetDefaultState(),Y=-1,ce=null,P.pop(),P.length>0?(v=P[P.length-1],J===!0&&Ee.setGlobalState(b.clippingPlanes,v.state.camera)):v=null,y.pop(),y.length>0?w=y[y.length-1]:w=null};function co(R,W,te,ne){if(R.visible===!1)return;if(R.layers.test(W.layers)){if(R.isGroup)te=R.renderOrder;else if(R.isLOD)R.autoUpdate===!0&&R.update(W);else if(R.isLight)v.pushLight(R),R.castShadow&&v.pushShadow(R);else if(R.isSprite){if(!R.frustumCulled||Ne.intersectsSprite(R)){ne&&Ge.setFromMatrixPosition(R.matrixWorld).applyMatrix4(Me);const De=de.update(R),Ae=R.material;Ae.visible&&w.push(R,De,Ae,te,Ge.z,null)}}else if((R.isMesh||R.isLine||R.isPoints)&&(!R.frustumCulled||Ne.intersectsObject(R))){const De=de.update(R),Ae=R.material;if(ne&&(R.boundingSphere!==void 0?(R.boundingSphere===null&&R.computeBoundingSphere(),Ge.copy(R.boundingSphere.center)):(De.boundingSphere===null&&De.computeBoundingSphere(),Ge.copy(De.boundingSphere.center)),Ge.applyMatrix4(R.matrixWorld).applyMatrix4(Me)),Array.isArray(Ae)){const We=De.groups;for(let Ke=0,Qe=We.length;Ke<Qe;Ke++){const je=We[Ke],Mt=Ae[je.materialIndex];Mt&&Mt.visible&&w.push(R,De,Mt,te,Ge.z,je)}}else Ae.visible&&w.push(R,De,Ae,te,Ge.z,null)}}const we=R.children;for(let De=0,Ae=we.length;De<Ae;De++)co(we[De],W,te,ne)}function Yi(R,W,te,ne){const j=R.opaque,we=R.transmissive,De=R.transparent;v.setupLightsView(te),J===!0&&Ee.setGlobalState(b.clippingPlanes,te),ne&&Ze.viewport(E.copy(ne)),j.length>0&&Mi(j,W,te),we.length>0&&Mi(we,W,te),De.length>0&&Mi(De,W,te),Ze.buffers.depth.setTest(!0),Ze.buffers.depth.setMask(!0),Ze.buffers.color.setMask(!0),Ze.setPolygonOffset(!1)}function Pr(R,W,te,ne){if((te.isScene===!0?te.overrideMaterial:null)!==null)return;v.state.transmissionRenderTarget[ne.id]===void 0&&(v.state.transmissionRenderTarget[ne.id]=new is(1,1,{generateMipmaps:!0,type:at.has("EXT_color_buffer_half_float")||at.has("EXT_color_buffer_float")?sa:Wi,minFilter:ts,samples:4,stencilBuffer:l,resolveDepthBuffer:!1,resolveStencilBuffer:!1,colorSpace:Tt.workingColorSpace}));const we=v.state.transmissionRenderTarget[ne.id],De=ne.viewport||E;we.setSize(De.z,De.w);const Ae=b.getRenderTarget();b.setRenderTarget(we),b.getClearColor(ee),ae=b.getClearAlpha(),ae<1&&b.setClearColor(16777215,.5),b.clear(),gt&&tt.render(te);const We=b.toneMapping;b.toneMapping=wr;const Ke=ne.viewport;if(ne.viewport!==void 0&&(ne.viewport=void 0),v.setupLightsView(ne),J===!0&&Ee.setGlobalState(b.clippingPlanes,ne),Mi(R,te,ne),N.updateMultisampleRenderTarget(we),N.updateRenderTargetMipmap(we),at.has("WEBGL_multisampled_render_to_texture")===!1){let Qe=!1;for(let je=0,Mt=W.length;je<Mt;je++){const Ct=W[je],bt=Ct.object,Ft=Ct.geometry,xt=Ct.material,ke=Ct.group;if(xt.side===zi&&bt.layers.test(ne.layers)){const qt=xt.side;xt.side=Ln,xt.needsUpdate=!0,ls(bt,te,ne,Ft,xt,ke),xt.side=qt,xt.needsUpdate=!0,Qe=!0}}Qe===!0&&(N.updateMultisampleRenderTarget(we),N.updateRenderTargetMipmap(we))}b.setRenderTarget(Ae),b.setClearColor(ee,ae),Ke!==void 0&&(ne.viewport=Ke),b.toneMapping=We}function Mi(R,W,te){const ne=W.isScene===!0?W.overrideMaterial:null;for(let j=0,we=R.length;j<we;j++){const De=R[j],Ae=De.object,We=De.geometry,Ke=ne===null?De.material:ne,Qe=De.group;Ae.layers.test(te.layers)&&ls(Ae,W,te,We,Ke,Qe)}}function ls(R,W,te,ne,j,we){R.onBeforeRender(b,W,te,ne,j,we),R.modelViewMatrix.multiplyMatrices(te.matrixWorldInverse,R.matrixWorld),R.normalMatrix.getNormalMatrix(R.modelViewMatrix),j.onBeforeRender(b,W,te,ne,R,we),j.transparent===!0&&j.side===zi&&j.forceSinglePass===!1?(j.side=Ln,j.needsUpdate=!0,b.renderBufferDirect(te,W,ne,j,R,we),j.side=Ar,j.needsUpdate=!0,b.renderBufferDirect(te,W,ne,j,R,we),j.side=zi):b.renderBufferDirect(te,W,ne,j,R,we),R.onAfterRender(b,W,te,ne,j,we)}function cs(R,W,te){W.isScene!==!0&&(W=dt);const ne=Je.get(R),j=v.state.lights,we=v.state.shadowsArray,De=j.state.version,Ae=Ye.getParameters(R,j.state,we,W,te),We=Ye.getProgramCacheKey(Ae);let Ke=ne.programs;ne.environment=R.isMeshStandardMaterial?W.environment:null,ne.fog=W.fog,ne.envMap=(R.isMeshStandardMaterial?K:A).get(R.envMap||ne.environment),ne.envMapRotation=ne.environment!==null&&R.envMap===null?W.environmentRotation:R.envMapRotation,Ke===void 0&&(R.addEventListener("dispose",ft),Ke=new Map,ne.programs=Ke);let Qe=Ke.get(We);if(Qe!==void 0){if(ne.currentProgram===Qe&&ne.lightsStateVersion===De)return ua(R,Ae),Qe}else Ae.uniforms=Ye.getUniforms(R),R.onBeforeCompile(Ae,b),Qe=Ye.acquireProgram(Ae,We),Ke.set(We,Qe),ne.uniforms=Ae.uniforms;const je=ne.uniforms;return(!R.isShaderMaterial&&!R.isRawShaderMaterial||R.clipping===!0)&&(je.clippingPlanes=Ee.uniform),ua(R,Ae),ne.needsLights=da(R),ne.lightsStateVersion=De,ne.needsLights&&(je.ambientLightColor.value=j.state.ambient,je.lightProbe.value=j.state.probe,je.directionalLights.value=j.state.directional,je.directionalLightShadows.value=j.state.directionalShadow,je.spotLights.value=j.state.spot,je.spotLightShadows.value=j.state.spotShadow,je.rectAreaLights.value=j.state.rectArea,je.ltc_1.value=j.state.rectAreaLTC1,je.ltc_2.value=j.state.rectAreaLTC2,je.pointLights.value=j.state.point,je.pointLightShadows.value=j.state.pointShadow,je.hemisphereLights.value=j.state.hemi,je.directionalShadowMap.value=j.state.directionalShadowMap,je.directionalShadowMatrix.value=j.state.directionalShadowMatrix,je.spotShadowMap.value=j.state.spotShadowMap,je.spotLightMatrix.value=j.state.spotLightMatrix,je.spotLightMap.value=j.state.spotLightMap,je.pointShadowMap.value=j.state.pointShadowMap,je.pointShadowMatrix.value=j.state.pointShadowMatrix),ne.currentProgram=Qe,ne.uniformsList=null,Qe}function ca(R){if(R.uniformsList===null){const W=R.currentProgram.getUniforms();R.uniformsList=jl.seqWithValue(W.seq,R.uniforms)}return R.uniformsList}function ua(R,W){const te=Je.get(R);te.outputColorSpace=W.outputColorSpace,te.batching=W.batching,te.batchingColor=W.batchingColor,te.instancing=W.instancing,te.instancingColor=W.instancingColor,te.instancingMorph=W.instancingMorph,te.skinning=W.skinning,te.morphTargets=W.morphTargets,te.morphNormals=W.morphNormals,te.morphColors=W.morphColors,te.morphTargetsCount=W.morphTargetsCount,te.numClippingPlanes=W.numClippingPlanes,te.numIntersection=W.numClipIntersection,te.vertexAlphas=W.vertexAlphas,te.vertexTangents=W.vertexTangents,te.toneMapping=W.toneMapping}function Ei(R,W,te,ne,j){W.isScene!==!0&&(W=dt),N.resetTextureUnits();const we=W.fog,De=ne.isMeshStandardMaterial?W.environment:null,Ae=U===null?b.outputColorSpace:U.isXRRenderTarget===!0?U.texture.colorSpace:br,We=(ne.isMeshStandardMaterial?K:A).get(ne.envMap||De),Ke=ne.vertexColors===!0&&!!te.attributes.color&&te.attributes.color.itemSize===4,Qe=!!te.attributes.tangent&&(!!ne.normalMap||ne.anisotropy>0),je=!!te.morphAttributes.position,Mt=!!te.morphAttributes.normal,Ct=!!te.morphAttributes.color;let bt=wr;ne.toneMapped&&(U===null||U.isXRRenderTarget===!0)&&(bt=b.toneMapping);const Ft=te.morphAttributes.position||te.morphAttributes.normal||te.morphAttributes.color,xt=Ft!==void 0?Ft.length:0,ke=Je.get(ne),qt=v.state.lights;if(J===!0&&(fe===!0||R!==ce)){const fn=R===ce&&ne.id===Y;Ee.setState(ne,R,fn)}let yt=!1;ne.version===ke.__version?(ke.needsLights&&ke.lightsStateVersion!==qt.state.version||ke.outputColorSpace!==Ae||j.isBatchedMesh&&ke.batching===!1||!j.isBatchedMesh&&ke.batching===!0||j.isBatchedMesh&&ke.batchingColor===!0&&j.colorTexture===null||j.isBatchedMesh&&ke.batchingColor===!1&&j.colorTexture!==null||j.isInstancedMesh&&ke.instancing===!1||!j.isInstancedMesh&&ke.instancing===!0||j.isSkinnedMesh&&ke.skinning===!1||!j.isSkinnedMesh&&ke.skinning===!0||j.isInstancedMesh&&ke.instancingColor===!0&&j.instanceColor===null||j.isInstancedMesh&&ke.instancingColor===!1&&j.instanceColor!==null||j.isInstancedMesh&&ke.instancingMorph===!0&&j.morphTexture===null||j.isInstancedMesh&&ke.instancingMorph===!1&&j.morphTexture!==null||ke.envMap!==We||ne.fog===!0&&ke.fog!==we||ke.numClippingPlanes!==void 0&&(ke.numClippingPlanes!==Ee.numPlanes||ke.numIntersection!==Ee.numIntersection)||ke.vertexAlphas!==Ke||ke.vertexTangents!==Qe||ke.morphTargets!==je||ke.morphNormals!==Mt||ke.morphColors!==Ct||ke.toneMapping!==bt||ke.morphTargetsCount!==xt)&&(yt=!0):(yt=!0,ke.__version=ne.version);let In=ke.currentProgram;yt===!0&&(In=cs(ne,W,j));let Jn=!1,tn=!1,Ti=!1;const Lt=In.getUniforms(),hi=ke.uniforms;if(Ze.useProgram(In.program)&&(Jn=!0,tn=!0,Ti=!0),ne.id!==Y&&(Y=ne.id,tn=!0),Jn||ce!==R){ht.reverseDepthBuffer?(Se.copy(R.projectionMatrix),oy(Se),ay(Se),Lt.setValue(z,"projectionMatrix",Se)):Lt.setValue(z,"projectionMatrix",R.projectionMatrix),Lt.setValue(z,"viewMatrix",R.matrixWorldInverse);const fn=Lt.map.cameraPosition;fn!==void 0&&fn.setValue(z,Pe.setFromMatrixPosition(R.matrixWorld)),ht.logarithmicDepthBuffer&&Lt.setValue(z,"logDepthBufFC",2/(Math.log(R.far+1)/Math.LN2)),(ne.isMeshPhongMaterial||ne.isMeshToonMaterial||ne.isMeshLambertMaterial||ne.isMeshBasicMaterial||ne.isMeshStandardMaterial||ne.isShaderMaterial)&&Lt.setValue(z,"isOrthographic",R.isOrthographicCamera===!0),ce!==R&&(ce=R,tn=!0,Ti=!0)}if(j.isSkinnedMesh){Lt.setOptional(z,j,"bindMatrix"),Lt.setOptional(z,j,"bindMatrixInverse");const fn=j.skeleton;fn&&(fn.boneTexture===null&&fn.computeBoneTexture(),Lt.setValue(z,"boneTexture",fn.boneTexture,N))}j.isBatchedMesh&&(Lt.setOptional(z,j,"batchingTexture"),Lt.setValue(z,"batchingTexture",j._matricesTexture,N),Lt.setOptional(z,j,"batchingIdTexture"),Lt.setValue(z,"batchingIdTexture",j._indirectTexture,N),Lt.setOptional(z,j,"batchingColorTexture"),j._colorsTexture!==null&&Lt.setValue(z,"batchingColorTexture",j._colorsTexture,N));const uo=te.morphAttributes;if((uo.position!==void 0||uo.normal!==void 0||uo.color!==void 0)&&et.update(j,te,In),(tn||ke.receiveShadow!==j.receiveShadow)&&(ke.receiveShadow=j.receiveShadow,Lt.setValue(z,"receiveShadow",j.receiveShadow)),ne.isMeshGouraudMaterial&&ne.envMap!==null&&(hi.envMap.value=We,hi.flipEnvMap.value=We.isCubeTexture&&We.isRenderTargetTexture===!1?-1:1),ne.isMeshStandardMaterial&&ne.envMap===null&&W.environment!==null&&(hi.envMapIntensity.value=W.environmentIntensity),tn&&(Lt.setValue(z,"toneMappingExposure",b.toneMappingExposure),ke.needsLights&&fa(hi,Ti),we&&ne.fog===!0&&Ce.refreshFogUniforms(hi,we),Ce.refreshMaterialUniforms(hi,ne,le,Z,v.state.transmissionRenderTarget[R.id]),jl.upload(z,ca(ke),hi,N)),ne.isShaderMaterial&&ne.uniformsNeedUpdate===!0&&(jl.upload(z,ca(ke),hi,N),ne.uniformsNeedUpdate=!1),ne.isSpriteMaterial&&Lt.setValue(z,"center",j.center),Lt.setValue(z,"modelViewMatrix",j.modelViewMatrix),Lt.setValue(z,"normalMatrix",j.normalMatrix),Lt.setValue(z,"modelMatrix",j.matrixWorld),ne.isShaderMaterial||ne.isRawShaderMaterial){const fn=ne.uniformsGroups;for(let us=0,fo=fn.length;us<fo;us++){const qi=fn[us];G.update(qi,In),G.bind(qi,In)}}return In}function fa(R,W){R.ambientLightColor.needsUpdate=W,R.lightProbe.needsUpdate=W,R.directionalLights.needsUpdate=W,R.directionalLightShadows.needsUpdate=W,R.pointLights.needsUpdate=W,R.pointLightShadows.needsUpdate=W,R.spotLights.needsUpdate=W,R.spotLightShadows.needsUpdate=W,R.rectAreaLights.needsUpdate=W,R.hemisphereLights.needsUpdate=W}function da(R){return R.isMeshLambertMaterial||R.isMeshToonMaterial||R.isMeshPhongMaterial||R.isMeshStandardMaterial||R.isShadowMaterial||R.isShaderMaterial&&R.lights===!0}this.getActiveCubeFace=function(){return V},this.getActiveMipmapLevel=function(){return O},this.getRenderTarget=function(){return U},this.setRenderTargetTextures=function(R,W,te){Je.get(R.texture).__webglTexture=W,Je.get(R.depthTexture).__webglTexture=te;const ne=Je.get(R);ne.__hasExternalTextures=!0,ne.__autoAllocateDepthBuffer=te===void 0,ne.__autoAllocateDepthBuffer||at.has("WEBGL_multisampled_render_to_texture")===!0&&(console.warn("THREE.WebGLRenderer: Render-to-texture extension was disabled because an external texture was provided"),ne.__useRenderToTexture=!1)},this.setRenderTargetFramebuffer=function(R,W){const te=Je.get(R);te.__webglFramebuffer=W,te.__useDefaultFramebuffer=W===void 0},this.setRenderTarget=function(R,W=0,te=0){U=R,V=W,O=te;let ne=!0,j=null,we=!1,De=!1;if(R){const We=Je.get(R);if(We.__useDefaultFramebuffer!==void 0)Ze.bindFramebuffer(z.FRAMEBUFFER,null),ne=!1;else if(We.__webglFramebuffer===void 0)N.setupRenderTarget(R);else if(We.__hasExternalTextures)N.rebindTextures(R,Je.get(R.texture).__webglTexture,Je.get(R.depthTexture).__webglTexture);else if(R.depthBuffer){const je=R.depthTexture;if(We.__boundDepthTexture!==je){if(je!==null&&Je.has(je)&&(R.width!==je.image.width||R.height!==je.image.height))throw new Error("WebGLRenderTarget: Attached DepthTexture is initialized to the incorrect size.");N.setupDepthRenderbuffer(R)}}const Ke=R.texture;(Ke.isData3DTexture||Ke.isDataArrayTexture||Ke.isCompressedArrayTexture)&&(De=!0);const Qe=Je.get(R).__webglFramebuffer;R.isWebGLCubeRenderTarget?(Array.isArray(Qe[W])?j=Qe[W][te]:j=Qe[W],we=!0):R.samples>0&&N.useMultisampledRTT(R)===!1?j=Je.get(R).__webglMultisampledFramebuffer:Array.isArray(Qe)?j=Qe[te]:j=Qe,E.copy(R.viewport),C.copy(R.scissor),re=R.scissorTest}else E.copy(L).multiplyScalar(le).floor(),C.copy(X).multiplyScalar(le).floor(),re=ve;if(Ze.bindFramebuffer(z.FRAMEBUFFER,j)&&ne&&Ze.drawBuffers(R,j),Ze.viewport(E),Ze.scissor(C),Ze.setScissorTest(re),we){const We=Je.get(R.texture);z.framebufferTexture2D(z.FRAMEBUFFER,z.COLOR_ATTACHMENT0,z.TEXTURE_CUBE_MAP_POSITIVE_X+W,We.__webglTexture,te)}else if(De){const We=Je.get(R.texture),Ke=W||0;z.framebufferTextureLayer(z.FRAMEBUFFER,z.COLOR_ATTACHMENT0,We.__webglTexture,te||0,Ke)}Y=-1},this.readRenderTargetPixels=function(R,W,te,ne,j,we,De){if(!(R&&R.isWebGLRenderTarget)){console.error("THREE.WebGLRenderer.readRenderTargetPixels: renderTarget is not THREE.WebGLRenderTarget.");return}let Ae=Je.get(R).__webglFramebuffer;if(R.isWebGLCubeRenderTarget&&De!==void 0&&(Ae=Ae[De]),Ae){Ze.bindFramebuffer(z.FRAMEBUFFER,Ae);try{const We=R.texture,Ke=We.format,Qe=We.type;if(!ht.textureFormatReadable(Ke)){console.error("THREE.WebGLRenderer.readRenderTargetPixels: renderTarget is not in RGBA or implementation defined format.");return}if(!ht.textureTypeReadable(Qe)){console.error("THREE.WebGLRenderer.readRenderTargetPixels: renderTarget is not in UnsignedByteType or implementation defined type.");return}W>=0&&W<=R.width-ne&&te>=0&&te<=R.height-j&&z.readPixels(W,te,ne,j,it.convert(Ke),it.convert(Qe),we)}finally{const We=U!==null?Je.get(U).__webglFramebuffer:null;Ze.bindFramebuffer(z.FRAMEBUFFER,We)}}},this.readRenderTargetPixelsAsync=async function(R,W,te,ne,j,we,De){if(!(R&&R.isWebGLRenderTarget))throw new Error("THREE.WebGLRenderer.readRenderTargetPixels: renderTarget is not THREE.WebGLRenderTarget.");let Ae=Je.get(R).__webglFramebuffer;if(R.isWebGLCubeRenderTarget&&De!==void 0&&(Ae=Ae[De]),Ae){const We=R.texture,Ke=We.format,Qe=We.type;if(!ht.textureFormatReadable(Ke))throw new Error("THREE.WebGLRenderer.readRenderTargetPixelsAsync: renderTarget is not in RGBA or implementation defined format.");if(!ht.textureTypeReadable(Qe))throw new Error("THREE.WebGLRenderer.readRenderTargetPixelsAsync: renderTarget is not in UnsignedByteType or implementation defined type.");if(W>=0&&W<=R.width-ne&&te>=0&&te<=R.height-j){Ze.bindFramebuffer(z.FRAMEBUFFER,Ae);const je=z.createBuffer();z.bindBuffer(z.PIXEL_PACK_BUFFER,je),z.bufferData(z.PIXEL_PACK_BUFFER,we.byteLength,z.STREAM_READ),z.readPixels(W,te,ne,j,it.convert(Ke),it.convert(Qe),0);const Mt=U!==null?Je.get(U).__webglFramebuffer:null;Ze.bindFramebuffer(z.FRAMEBUFFER,Mt);const Ct=z.fenceSync(z.SYNC_GPU_COMMANDS_COMPLETE,0);return z.flush(),await sy(z,Ct,4),z.bindBuffer(z.PIXEL_PACK_BUFFER,je),z.getBufferSubData(z.PIXEL_PACK_BUFFER,0,we),z.deleteBuffer(je),z.deleteSync(Ct),we}else throw new Error("THREE.WebGLRenderer.readRenderTargetPixelsAsync: requested read bounds are out of range.")}},this.copyFramebufferToTexture=function(R,W=null,te=0){R.isTexture!==!0&&(Wl("WebGLRenderer: copyFramebufferToTexture function signature has changed."),W=arguments[0]||null,R=arguments[1]);const ne=Math.pow(2,-te),j=Math.floor(R.image.width*ne),we=Math.floor(R.image.height*ne),De=W!==null?W.x:0,Ae=W!==null?W.y:0;N.setTexture2D(R,0),z.copyTexSubImage2D(z.TEXTURE_2D,te,0,0,De,Ae,j,we),Ze.unbindTexture()},this.copyTextureToTexture=function(R,W,te=null,ne=null,j=0){R.isTexture!==!0&&(Wl("WebGLRenderer: copyTextureToTexture function signature has changed."),ne=arguments[0]||null,R=arguments[1],W=arguments[2],j=arguments[3]||0,te=null);let we,De,Ae,We,Ke,Qe;te!==null?(we=te.max.x-te.min.x,De=te.max.y-te.min.y,Ae=te.min.x,We=te.min.y):(we=R.image.width,De=R.image.height,Ae=0,We=0),ne!==null?(Ke=ne.x,Qe=ne.y):(Ke=0,Qe=0);const je=it.convert(W.format),Mt=it.convert(W.type);N.setTexture2D(W,0),z.pixelStorei(z.UNPACK_FLIP_Y_WEBGL,W.flipY),z.pixelStorei(z.UNPACK_PREMULTIPLY_ALPHA_WEBGL,W.premultiplyAlpha),z.pixelStorei(z.UNPACK_ALIGNMENT,W.unpackAlignment);const Ct=z.getParameter(z.UNPACK_ROW_LENGTH),bt=z.getParameter(z.UNPACK_IMAGE_HEIGHT),Ft=z.getParameter(z.UNPACK_SKIP_PIXELS),xt=z.getParameter(z.UNPACK_SKIP_ROWS),ke=z.getParameter(z.UNPACK_SKIP_IMAGES),qt=R.isCompressedTexture?R.mipmaps[j]:R.image;z.pixelStorei(z.UNPACK_ROW_LENGTH,qt.width),z.pixelStorei(z.UNPACK_IMAGE_HEIGHT,qt.height),z.pixelStorei(z.UNPACK_SKIP_PIXELS,Ae),z.pixelStorei(z.UNPACK_SKIP_ROWS,We),R.isDataTexture?z.texSubImage2D(z.TEXTURE_2D,j,Ke,Qe,we,De,je,Mt,qt.data):R.isCompressedTexture?z.compressedTexSubImage2D(z.TEXTURE_2D,j,Ke,Qe,qt.width,qt.height,je,qt.data):z.texSubImage2D(z.TEXTURE_2D,j,Ke,Qe,we,De,je,Mt,qt),z.pixelStorei(z.UNPACK_ROW_LENGTH,Ct),z.pixelStorei(z.UNPACK_IMAGE_HEIGHT,bt),z.pixelStorei(z.UNPACK_SKIP_PIXELS,Ft),z.pixelStorei(z.UNPACK_SKIP_ROWS,xt),z.pixelStorei(z.UNPACK_SKIP_IMAGES,ke),j===0&&W.generateMipmaps&&z.generateMipmap(z.TEXTURE_2D),Ze.unbindTexture()},this.copyTextureToTexture3D=function(R,W,te=null,ne=null,j=0){R.isTexture!==!0&&(Wl("WebGLRenderer: copyTextureToTexture3D function signature has changed."),te=arguments[0]||null,ne=arguments[1]||null,R=arguments[2],W=arguments[3],j=arguments[4]||0);let we,De,Ae,We,Ke,Qe,je,Mt,Ct;const bt=R.isCompressedTexture?R.mipmaps[j]:R.image;te!==null?(we=te.max.x-te.min.x,De=te.max.y-te.min.y,Ae=te.max.z-te.min.z,We=te.min.x,Ke=te.min.y,Qe=te.min.z):(we=bt.width,De=bt.height,Ae=bt.depth,We=0,Ke=0,Qe=0),ne!==null?(je=ne.x,Mt=ne.y,Ct=ne.z):(je=0,Mt=0,Ct=0);const Ft=it.convert(W.format),xt=it.convert(W.type);let ke;if(W.isData3DTexture)N.setTexture3D(W,0),ke=z.TEXTURE_3D;else if(W.isDataArrayTexture||W.isCompressedArrayTexture)N.setTexture2DArray(W,0),ke=z.TEXTURE_2D_ARRAY;else{console.warn("THREE.WebGLRenderer.copyTextureToTexture3D: only supports THREE.DataTexture3D and THREE.DataTexture2DArray.");return}z.pixelStorei(z.UNPACK_FLIP_Y_WEBGL,W.flipY),z.pixelStorei(z.UNPACK_PREMULTIPLY_ALPHA_WEBGL,W.premultiplyAlpha),z.pixelStorei(z.UNPACK_ALIGNMENT,W.unpackAlignment);const qt=z.getParameter(z.UNPACK_ROW_LENGTH),yt=z.getParameter(z.UNPACK_IMAGE_HEIGHT),In=z.getParameter(z.UNPACK_SKIP_PIXELS),Jn=z.getParameter(z.UNPACK_SKIP_ROWS),tn=z.getParameter(z.UNPACK_SKIP_IMAGES);z.pixelStorei(z.UNPACK_ROW_LENGTH,bt.width),z.pixelStorei(z.UNPACK_IMAGE_HEIGHT,bt.height),z.pixelStorei(z.UNPACK_SKIP_PIXELS,We),z.pixelStorei(z.UNPACK_SKIP_ROWS,Ke),z.pixelStorei(z.UNPACK_SKIP_IMAGES,Qe),R.isDataTexture||R.isData3DTexture?z.texSubImage3D(ke,j,je,Mt,Ct,we,De,Ae,Ft,xt,bt.data):W.isCompressedArrayTexture?z.compressedTexSubImage3D(ke,j,je,Mt,Ct,we,De,Ae,Ft,bt.data):z.texSubImage3D(ke,j,je,Mt,Ct,we,De,Ae,Ft,xt,bt),z.pixelStorei(z.UNPACK_ROW_LENGTH,qt),z.pixelStorei(z.UNPACK_IMAGE_HEIGHT,yt),z.pixelStorei(z.UNPACK_SKIP_PIXELS,In),z.pixelStorei(z.UNPACK_SKIP_ROWS,Jn),z.pixelStorei(z.UNPACK_SKIP_IMAGES,tn),j===0&&W.generateMipmaps&&z.generateMipmap(ke),Ze.unbindTexture()},this.initRenderTarget=function(R){Je.get(R).__webglFramebuffer===void 0&&N.setupRenderTarget(R)},this.initTexture=function(R){R.isCubeTexture?N.setTextureCube(R,0):R.isData3DTexture?N.setTexture3D(R,0):R.isDataArrayTexture||R.isCompressedArrayTexture?N.setTexture2DArray(R,0):N.setTexture2D(R,0),Ze.unbindTexture()},this.resetState=function(){V=0,O=0,U=null,Ze.reset(),Et.reset()},typeof __THREE_DEVTOOLS__<"u"&&__THREE_DEVTOOLS__.dispatchEvent(new CustomEvent("observe",{detail:this}))}get coordinateSystem(){return Vi}get outputColorSpace(){return this._outputColorSpace}set outputColorSpace(e){this._outputColorSpace=e;const t=this.getContext();t.drawingBufferColorSpace=e===Ed?"display-p3":"srgb",t.unpackColorSpace=Tt.workingColorSpace===nc?"display-p3":"srgb"}}class vw extends vn{constructor(){super(),this.isScene=!0,this.type="Scene",this.background=null,this.environment=null,this.fog=null,this.backgroundBlurriness=0,this.backgroundIntensity=1,this.backgroundRotation=new Si,this.environmentIntensity=1,this.environmentRotation=new Si,this.overrideMaterial=null,typeof __THREE_DEVTOOLS__<"u"&&__THREE_DEVTOOLS__.dispatchEvent(new CustomEvent("observe",{detail:this}))}copy(e,t){return super.copy(e,t),e.background!==null&&(this.background=e.background.clone()),e.environment!==null&&(this.environment=e.environment.clone()),e.fog!==null&&(this.fog=e.fog.clone()),this.backgroundBlurriness=e.backgroundBlurriness,this.backgroundIntensity=e.backgroundIntensity,this.backgroundRotation.copy(e.backgroundRotation),this.environmentIntensity=e.environmentIntensity,this.environmentRotation.copy(e.environmentRotation),e.overrideMaterial!==null&&(this.overrideMaterial=e.overrideMaterial.clone()),this.matrixAutoUpdate=e.matrixAutoUpdate,this}toJSON(e){const t=super.toJSON(e);return this.fog!==null&&(t.object.fog=this.fog.toJSON()),this.backgroundBlurriness>0&&(t.object.backgroundBlurriness=this.backgroundBlurriness),this.backgroundIntensity!==1&&(t.object.backgroundIntensity=this.backgroundIntensity),t.object.backgroundRotation=this.backgroundRotation.toArray(),this.environmentIntensity!==1&&(t.object.environmentIntensity=this.environmentIntensity),t.object.environmentRotation=this.environmentRotation.toArray(),t}}class xw extends aa{constructor(e){super(),this.isMeshStandardMaterial=!0,this.defines={STANDARD:""},this.type="MeshStandardMaterial",this.color=new vt(16777215),this.roughness=1,this.metalness=0,this.map=null,this.lightMap=null,this.lightMapIntensity=1,this.aoMap=null,this.aoMapIntensity=1,this.emissive=new vt(0),this.emissiveIntensity=1,this.emissiveMap=null,this.bumpMap=null,this.bumpScale=1,this.normalMap=null,this.normalMapType=jg,this.normalScale=new rt(1,1),this.displacementMap=null,this.displacementScale=1,this.displacementBias=0,this.roughnessMap=null,this.metalnessMap=null,this.alphaMap=null,this.envMap=null,this.envMapRotation=new Si,this.envMapIntensity=1,this.wireframe=!1,this.wireframeLinewidth=1,this.wireframeLinecap="round",this.wireframeLinejoin="round",this.flatShading=!1,this.fog=!0,this.setValues(e)}copy(e){return super.copy(e),this.defines={STANDARD:""},this.color.copy(e.color),this.roughness=e.roughness,this.metalness=e.metalness,this.map=e.map,this.lightMap=e.lightMap,this.lightMapIntensity=e.lightMapIntensity,this.aoMap=e.aoMap,this.aoMapIntensity=e.aoMapIntensity,this.emissive.copy(e.emissive),this.emissiveMap=e.emissiveMap,this.emissiveIntensity=e.emissiveIntensity,this.bumpMap=e.bumpMap,this.bumpScale=e.bumpScale,this.normalMap=e.normalMap,this.normalMapType=e.normalMapType,this.normalScale.copy(e.normalScale),this.displacementMap=e.displacementMap,this.displacementScale=e.displacementScale,this.displacementBias=e.displacementBias,this.roughnessMap=e.roughnessMap,this.metalnessMap=e.metalnessMap,this.alphaMap=e.alphaMap,this.envMap=e.envMap,this.envMapRotation.copy(e.envMapRotation),this.envMapIntensity=e.envMapIntensity,this.wireframe=e.wireframe,this.wireframeLinewidth=e.wireframeLinewidth,this.wireframeLinecap=e.wireframeLinecap,this.wireframeLinejoin=e.wireframeLinejoin,this.flatShading=e.flatShading,this.fog=e.fog,this}}const ug={enabled:!1,files:{},add:function(r,e){this.enabled!==!1&&(this.files[r]=e)},get:function(r){if(this.enabled!==!1)return this.files[r]},remove:function(r){delete this.files[r]},clear:function(){this.files={}}};class yw{constructor(e,t,s){const a=this;let l=!1,u=0,f=0,h;const p=[];this.onStart=void 0,this.onLoad=e,this.onProgress=t,this.onError=s,this.itemStart=function(m){f++,l===!1&&a.onStart!==void 0&&a.onStart(m,u,f),l=!0},this.itemEnd=function(m){u++,a.onProgress!==void 0&&a.onProgress(m,u,f),u===f&&(l=!1,a.onLoad!==void 0&&a.onLoad())},this.itemError=function(m){a.onError!==void 0&&a.onError(m)},this.resolveURL=function(m){return h?h(m):m},this.setURLModifier=function(m){return h=m,this},this.addHandler=function(m,_){return p.push(m,_),this},this.removeHandler=function(m){const _=p.indexOf(m);return _!==-1&&p.splice(_,2),this},this.getHandler=function(m){for(let _=0,x=p.length;_<x;_+=2){const S=p[_],T=p[_+1];if(S.global&&(S.lastIndex=0),S.test(m))return T}return null}}}const Sw=new yw;class Cd{constructor(e){this.manager=e!==void 0?e:Sw,this.crossOrigin="anonymous",this.withCredentials=!1,this.path="",this.resourcePath="",this.requestHeader={}}load(){}loadAsync(e,t){const s=this;return new Promise(function(a,l){s.load(e,a,t,l)})}parse(){}setCrossOrigin(e){return this.crossOrigin=e,this}setWithCredentials(e){return this.withCredentials=e,this}setPath(e){return this.path=e,this}setResourcePath(e){return this.resourcePath=e,this}setRequestHeader(e){return this.requestHeader=e,this}}Cd.DEFAULT_MATERIAL_NAME="__DEFAULT";const ki={};class Mw extends Error{constructor(e,t){super(e),this.response=t}}class Ew extends Cd{constructor(e){super(e)}load(e,t,s,a){e===void 0&&(e=""),this.path!==void 0&&(e=this.path+e),e=this.manager.resolveURL(e);const l=ug.get(e);if(l!==void 0)return this.manager.itemStart(e),setTimeout(()=>{t&&t(l),this.manager.itemEnd(e)},0),l;if(ki[e]!==void 0){ki[e].push({onLoad:t,onProgress:s,onError:a});return}ki[e]=[],ki[e].push({onLoad:t,onProgress:s,onError:a});const u=new Request(e,{headers:new Headers(this.requestHeader),credentials:this.withCredentials?"include":"same-origin"}),f=this.mimeType,h=this.responseType;fetch(u).then(p=>{if(p.status===200||p.status===0){if(p.status===0&&console.warn("THREE.FileLoader: HTTP Status 0 received."),typeof ReadableStream>"u"||p.body===void 0||p.body.getReader===void 0)return p;const m=ki[e],_=p.body.getReader(),x=p.headers.get("X-File-Size")||p.headers.get("Content-Length"),S=x?parseInt(x):0,T=S!==0;let w=0;const v=new ReadableStream({start(y){P();function P(){_.read().then(({done:b,value:D})=>{if(b)y.close();else{w+=D.byteLength;const V=new ProgressEvent("progress",{lengthComputable:T,loaded:w,total:S});for(let O=0,U=m.length;O<U;O++){const Y=m[O];Y.onProgress&&Y.onProgress(V)}y.enqueue(D),P()}},b=>{y.error(b)})}}});return new Response(v)}else throw new Mw(`fetch for "${p.url}" responded with ${p.status}: ${p.statusText}`,p)}).then(p=>{switch(h){case"arraybuffer":return p.arrayBuffer();case"blob":return p.blob();case"document":return p.text().then(m=>new DOMParser().parseFromString(m,f));case"json":return p.json();default:if(f===void 0)return p.text();{const _=/charset="?([^;"\s]*)"?/i.exec(f),x=_&&_[1]?_[1].toLowerCase():void 0,S=new TextDecoder(x);return p.arrayBuffer().then(T=>S.decode(T))}}}).then(p=>{ug.add(e,p);const m=ki[e];delete ki[e];for(let _=0,x=m.length;_<x;_++){const S=m[_];S.onLoad&&S.onLoad(p)}}).catch(p=>{const m=ki[e];if(m===void 0)throw this.manager.itemError(e),p;delete ki[e];for(let _=0,x=m.length;_<x;_++){const S=m[_];S.onError&&S.onError(p)}this.manager.itemError(e)}).finally(()=>{this.manager.itemEnd(e)}),this.manager.itemStart(e)}setResponseType(e){return this.responseType=e,this}setMimeType(e){return this.mimeType=e,this}}class d_ extends vn{constructor(e,t=1){super(),this.isLight=!0,this.type="Light",this.color=new vt(e),this.intensity=t}dispose(){}copy(e,t){return super.copy(e,t),this.color.copy(e.color),this.intensity=e.intensity,this}toJSON(e){const t=super.toJSON(e);return t.object.color=this.color.getHex(),t.object.intensity=this.intensity,this.groundColor!==void 0&&(t.object.groundColor=this.groundColor.getHex()),this.distance!==void 0&&(t.object.distance=this.distance),this.angle!==void 0&&(t.object.angle=this.angle),this.decay!==void 0&&(t.object.decay=this.decay),this.penumbra!==void 0&&(t.object.penumbra=this.penumbra),this.shadow!==void 0&&(t.object.shadow=this.shadow.toJSON()),this.target!==void 0&&(t.object.target=this.target.uuid),t}}const xf=new Gt,fg=new Q,dg=new Q;class Tw{constructor(e){this.camera=e,this.intensity=1,this.bias=0,this.normalBias=0,this.radius=1,this.blurSamples=8,this.mapSize=new rt(512,512),this.map=null,this.mapPass=null,this.matrix=new Gt,this.autoUpdate=!0,this.needsUpdate=!1,this._frustum=new wd,this._frameExtents=new rt(1,1),this._viewportCount=1,this._viewports=[new Vt(0,0,1,1)]}getViewportCount(){return this._viewportCount}getFrustum(){return this._frustum}updateMatrices(e){const t=this.camera,s=this.matrix;fg.setFromMatrixPosition(e.matrixWorld),t.position.copy(fg),dg.setFromMatrixPosition(e.target.matrixWorld),t.lookAt(dg),t.updateMatrixWorld(),xf.multiplyMatrices(t.projectionMatrix,t.matrixWorldInverse),this._frustum.setFromProjectionMatrix(xf),s.set(.5,0,0,.5,0,.5,0,.5,0,0,.5,.5,0,0,0,1),s.multiply(xf)}getViewport(e){return this._viewports[e]}getFrameExtents(){return this._frameExtents}dispose(){this.map&&this.map.dispose(),this.mapPass&&this.mapPass.dispose()}copy(e){return this.camera=e.camera.clone(),this.intensity=e.intensity,this.bias=e.bias,this.radius=e.radius,this.mapSize.copy(e.mapSize),this}clone(){return new this.constructor().copy(this)}toJSON(){const e={};return this.intensity!==1&&(e.intensity=this.intensity),this.bias!==0&&(e.bias=this.bias),this.normalBias!==0&&(e.normalBias=this.normalBias),this.radius!==1&&(e.radius=this.radius),(this.mapSize.x!==512||this.mapSize.y!==512)&&(e.mapSize=this.mapSize.toArray()),e.camera=this.camera.toJSON(!1).object,delete e.camera.matrix,e}}class ww extends Tw{constructor(){super(new o_(-5,5,5,-5,.5,500)),this.isDirectionalLightShadow=!0}}class Aw extends d_{constructor(e,t){super(e,t),this.isDirectionalLight=!0,this.type="DirectionalLight",this.position.copy(vn.DEFAULT_UP),this.updateMatrix(),this.target=new vn,this.shadow=new ww}dispose(){this.shadow.dispose()}copy(e){return super.copy(e),this.target=e.target.clone(),this.shadow=e.shadow.clone(),this}}class Cw extends d_{constructor(e,t){super(e,t),this.isAmbientLight=!0,this.type="AmbientLight"}}class hg{constructor(e=1,t=0,s=0){return this.radius=e,this.phi=t,this.theta=s,this}set(e,t,s){return this.radius=e,this.phi=t,this.theta=s,this}copy(e){return this.radius=e.radius,this.phi=e.phi,this.theta=e.theta,this}makeSafe(){return this.phi=Math.max(1e-6,Math.min(Math.PI-1e-6,this.phi)),this}setFromVector3(e){return this.setFromCartesianCoords(e.x,e.y,e.z)}setFromCartesianCoords(e,t,s){return this.radius=Math.sqrt(e*e+t*t+s*s),this.radius===0?(this.theta=0,this.phi=0):(this.theta=Math.atan2(e,s),this.phi=Math.acos(Mn(t/this.radius,-1,1))),this}clone(){return new this.constructor().copy(this)}}class Rw extends os{constructor(e,t=null){super(),this.object=e,this.domElement=t,this.enabled=!0,this.state=-1,this.keys={},this.mouseButtons={LEFT:null,MIDDLE:null,RIGHT:null},this.touches={ONE:null,TWO:null}}connect(){}disconnect(){}dispose(){}update(){}}typeof __THREE_DEVTOOLS__<"u"&&__THREE_DEVTOOLS__.dispatchEvent(new CustomEvent("register",{detail:{revision:gd}}));typeof window<"u"&&(window.__THREE__?console.warn("WARNING: Multiple instances of Three.js being imported."):window.__THREE__=gd);class bw extends Cd{constructor(e){super(e)}load(e,t,s,a){const l=this,u=new Ew(this.manager);u.setPath(this.path),u.setResponseType("arraybuffer"),u.setRequestHeader(this.requestHeader),u.setWithCredentials(this.withCredentials),u.load(e,function(f){try{t(l.parse(f))}catch(h){a?a(h):console.error(h),l.manager.itemError(e)}},s,a)}parse(e){function t(p){const m=new DataView(p),_=32/8*3+32/8*3*3+16/8,x=m.getUint32(80,!0);if(80+32/8+x*_===m.byteLength)return!0;const T=[115,111,108,105,100];for(let w=0;w<5;w++)if(s(T,m,w))return!1;return!0}function s(p,m,_){for(let x=0,S=p.length;x<S;x++)if(p[x]!==m.getUint8(_+x))return!1;return!0}function a(p){const m=new DataView(p),_=m.getUint32(80,!0);let x,S,T,w=!1,v,y,P,b,D;for(let C=0;C<70;C++)m.getUint32(C,!1)==1129270351&&m.getUint8(C+4)==82&&m.getUint8(C+5)==61&&(w=!0,v=new Float32Array(_*3*3),y=m.getUint8(C+6)/255,P=m.getUint8(C+7)/255,b=m.getUint8(C+8)/255,D=m.getUint8(C+9)/255);const V=84,O=50,U=new ji,Y=new Float32Array(_*3*3),ce=new Float32Array(_*3*3),E=new vt;for(let C=0;C<_;C++){const re=V+C*O,ee=m.getFloat32(re,!0),ae=m.getFloat32(re+4,!0),ue=m.getFloat32(re+8,!0);if(w){const Z=m.getUint16(re+48,!0);(Z&32768)===0?(x=(Z&31)/31,S=(Z>>5&31)/31,T=(Z>>10&31)/31):(x=y,S=P,T=b)}for(let Z=1;Z<=3;Z++){const le=re+Z*12,F=C*3*3+(Z-1)*3;Y[F]=m.getFloat32(le,!0),Y[F+1]=m.getFloat32(le+4,!0),Y[F+2]=m.getFloat32(le+8,!0),ce[F]=ee,ce[F+1]=ae,ce[F+2]=ue,w&&(E.setRGB(x,S,T,ci),v[F]=E.r,v[F+1]=E.g,v[F+2]=E.b)}}return U.setAttribute("position",new Vn(Y,3)),U.setAttribute("normal",new Vn(ce,3)),w&&(U.setAttribute("color",new Vn(v,3)),U.hasColors=!0,U.alpha=D),U}function l(p){const m=new ji,_=/solid([\s\S]*?)endsolid/g,x=/facet([\s\S]*?)endfacet/g,S=/solid\s(.+)/;let T=0;const w=/[\s]+([+-]?(?:\d*)(?:\.\d*)?(?:[eE][+-]?\d+)?)/.source,v=new RegExp("vertex"+w+w+w,"g"),y=new RegExp("normal"+w+w+w,"g"),P=[],b=[],D=[],V=new Q;let O,U=0,Y=0,ce=0;for(;(O=_.exec(p))!==null;){Y=ce;const E=O[0],C=(O=S.exec(E))!==null?O[1]:"";for(D.push(C);(O=x.exec(E))!==null;){let ae=0,ue=0;const Z=O[0];for(;(O=y.exec(Z))!==null;)V.x=parseFloat(O[1]),V.y=parseFloat(O[2]),V.z=parseFloat(O[3]),ue++;for(;(O=v.exec(Z))!==null;)P.push(parseFloat(O[1]),parseFloat(O[2]),parseFloat(O[3])),b.push(V.x,V.y,V.z),ae++,ce++;ue!==1&&console.error("THREE.STLLoader: Something isn't right with the normal of face number "+T),ae!==3&&console.error("THREE.STLLoader: Something isn't right with the vertices of face number "+T),T++}const re=Y,ee=ce-Y;m.userData.groupNames=D,m.addGroup(re,ee,U),U++}return m.setAttribute("position",new Gi(P,3)),m.setAttribute("normal",new Gi(b,3)),m}function u(p){return typeof p!="string"?new TextDecoder().decode(p):p}function f(p){if(typeof p=="string"){const m=new Uint8Array(p.length);for(let _=0;_<p.length;_++)m[_]=p.charCodeAt(_)&255;return m.buffer||m}else return p}const h=f(e);return t(h)?a(h):l(u(e))}}const pg={type:"change"},Rd={type:"start"},h_={type:"end"},Ol=new Kg,mg=new yr,Pw=Math.cos(70*iy.DEG2RAD),Zt=new Q,Pn=2*Math.PI,Rt={NONE:-1,ROTATE:0,DOLLY:1,PAN:2,TOUCH_ROTATE:3,TOUCH_PAN:4,TOUCH_DOLLY_PAN:5,TOUCH_DOLLY_ROTATE:6},yf=1e-6;class Lw extends Rw{constructor(e,t=null){super(e,t),this.state=Rt.NONE,this.enabled=!0,this.target=new Q,this.cursor=new Q,this.minDistance=0,this.maxDistance=1/0,this.minZoom=0,this.maxZoom=1/0,this.minTargetRadius=0,this.maxTargetRadius=1/0,this.minPolarAngle=0,this.maxPolarAngle=Math.PI,this.minAzimuthAngle=-1/0,this.maxAzimuthAngle=1/0,this.enableDamping=!1,this.dampingFactor=.05,this.enableZoom=!0,this.zoomSpeed=1,this.enableRotate=!0,this.rotateSpeed=1,this.enablePan=!0,this.panSpeed=1,this.screenSpacePanning=!0,this.keyPanSpeed=7,this.zoomToCursor=!1,this.autoRotate=!1,this.autoRotateSpeed=2,this.keys={LEFT:"ArrowLeft",UP:"ArrowUp",RIGHT:"ArrowRight",BOTTOM:"ArrowDown"},this.mouseButtons={LEFT:Ks.ROTATE,MIDDLE:Ks.DOLLY,RIGHT:Ks.PAN},this.touches={ONE:qs.ROTATE,TWO:qs.DOLLY_PAN},this.target0=this.target.clone(),this.position0=this.object.position.clone(),this.zoom0=this.object.zoom,this._domElementKeyEvents=null,this._lastPosition=new Q,this._lastQuaternion=new rs,this._lastTargetPosition=new Q,this._quat=new rs().setFromUnitVectors(e.up,new Q(0,1,0)),this._quatInverse=this._quat.clone().invert(),this._spherical=new hg,this._sphericalDelta=new hg,this._scale=1,this._panOffset=new Q,this._rotateStart=new rt,this._rotateEnd=new rt,this._rotateDelta=new rt,this._panStart=new rt,this._panEnd=new rt,this._panDelta=new rt,this._dollyStart=new rt,this._dollyEnd=new rt,this._dollyDelta=new rt,this._dollyDirection=new Q,this._mouse=new rt,this._performCursorZoom=!1,this._pointers=[],this._pointerPositions={},this._controlActive=!1,this._onPointerMove=Nw.bind(this),this._onPointerDown=Dw.bind(this),this._onPointerUp=Iw.bind(this),this._onContextMenu=Hw.bind(this),this._onMouseWheel=Ow.bind(this),this._onKeyDown=kw.bind(this),this._onTouchStart=Bw.bind(this),this._onTouchMove=zw.bind(this),this._onMouseDown=Uw.bind(this),this._onMouseMove=Fw.bind(this),this._interceptControlDown=Vw.bind(this),this._interceptControlUp=Gw.bind(this),this.domElement!==null&&this.connect(),this.update()}connect(){this.domElement.addEventListener("pointerdown",this._onPointerDown),this.domElement.addEventListener("pointercancel",this._onPointerUp),this.domElement.addEventListener("contextmenu",this._onContextMenu),this.domElement.addEventListener("wheel",this._onMouseWheel,{passive:!1}),this.domElement.getRootNode().addEventListener("keydown",this._interceptControlDown,{passive:!0,capture:!0}),this.domElement.style.touchAction="none"}disconnect(){this.domElement.removeEventListener("pointerdown",this._onPointerDown),this.domElement.removeEventListener("pointermove",this._onPointerMove),this.domElement.removeEventListener("pointerup",this._onPointerUp),this.domElement.removeEventListener("pointercancel",this._onPointerUp),this.domElement.removeEventListener("wheel",this._onMouseWheel),this.domElement.removeEventListener("contextmenu",this._onContextMenu),this.stopListenToKeyEvents(),this.domElement.getRootNode().removeEventListener("keydown",this._interceptControlDown,{capture:!0}),this.domElement.style.touchAction="auto"}dispose(){this.disconnect()}getPolarAngle(){return this._spherical.phi}getAzimuthalAngle(){return this._spherical.theta}getDistance(){return this.object.position.distanceTo(this.target)}listenToKeyEvents(e){e.addEventListener("keydown",this._onKeyDown),this._domElementKeyEvents=e}stopListenToKeyEvents(){this._domElementKeyEvents!==null&&(this._domElementKeyEvents.removeEventListener("keydown",this._onKeyDown),this._domElementKeyEvents=null)}saveState(){this.target0.copy(this.target),this.position0.copy(this.object.position),this.zoom0=this.object.zoom}reset(){this.target.copy(this.target0),this.object.position.copy(this.position0),this.object.zoom=this.zoom0,this.object.updateProjectionMatrix(),this.dispatchEvent(pg),this.update(),this.state=Rt.NONE}update(e=null){const t=this.object.position;Zt.copy(t).sub(this.target),Zt.applyQuaternion(this._quat),this._spherical.setFromVector3(Zt),this.autoRotate&&this.state===Rt.NONE&&this._rotateLeft(this._getAutoRotationAngle(e)),this.enableDamping?(this._spherical.theta+=this._sphericalDelta.theta*this.dampingFactor,this._spherical.phi+=this._sphericalDelta.phi*this.dampingFactor):(this._spherical.theta+=this._sphericalDelta.theta,this._spherical.phi+=this._sphericalDelta.phi);let s=this.minAzimuthAngle,a=this.maxAzimuthAngle;isFinite(s)&&isFinite(a)&&(s<-Math.PI?s+=Pn:s>Math.PI&&(s-=Pn),a<-Math.PI?a+=Pn:a>Math.PI&&(a-=Pn),s<=a?this._spherical.theta=Math.max(s,Math.min(a,this._spherical.theta)):this._spherical.theta=this._spherical.theta>(s+a)/2?Math.max(s,this._spherical.theta):Math.min(a,this._spherical.theta)),this._spherical.phi=Math.max(this.minPolarAngle,Math.min(this.maxPolarAngle,this._spherical.phi)),this._spherical.makeSafe(),this.enableDamping===!0?this.target.addScaledVector(this._panOffset,this.dampingFactor):this.target.add(this._panOffset),this.target.sub(this.cursor),this.target.clampLength(this.minTargetRadius,this.maxTargetRadius),this.target.add(this.cursor);let l=!1;if(this.zoomToCursor&&this._performCursorZoom||this.object.isOrthographicCamera)this._spherical.radius=this._clampDistance(this._spherical.radius);else{const u=this._spherical.radius;this._spherical.radius=this._clampDistance(this._spherical.radius*this._scale),l=u!=this._spherical.radius}if(Zt.setFromSpherical(this._spherical),Zt.applyQuaternion(this._quatInverse),t.copy(this.target).add(Zt),this.object.lookAt(this.target),this.enableDamping===!0?(this._sphericalDelta.theta*=1-this.dampingFactor,this._sphericalDelta.phi*=1-this.dampingFactor,this._panOffset.multiplyScalar(1-this.dampingFactor)):(this._sphericalDelta.set(0,0,0),this._panOffset.set(0,0,0)),this.zoomToCursor&&this._performCursorZoom){let u=null;if(this.object.isPerspectiveCamera){const f=Zt.length();u=this._clampDistance(f*this._scale);const h=f-u;this.object.position.addScaledVector(this._dollyDirection,h),this.object.updateMatrixWorld(),l=!!h}else if(this.object.isOrthographicCamera){const f=new Q(this._mouse.x,this._mouse.y,0);f.unproject(this.object);const h=this.object.zoom;this.object.zoom=Math.max(this.minZoom,Math.min(this.maxZoom,this.object.zoom/this._scale)),this.object.updateProjectionMatrix(),l=h!==this.object.zoom;const p=new Q(this._mouse.x,this._mouse.y,0);p.unproject(this.object),this.object.position.sub(p).add(f),this.object.updateMatrixWorld(),u=Zt.length()}else console.warn("WARNING: OrbitControls.js encountered an unknown camera type - zoom to cursor disabled."),this.zoomToCursor=!1;u!==null&&(this.screenSpacePanning?this.target.set(0,0,-1).transformDirection(this.object.matrix).multiplyScalar(u).add(this.object.position):(Ol.origin.copy(this.object.position),Ol.direction.set(0,0,-1).transformDirection(this.object.matrix),Math.abs(this.object.up.dot(Ol.direction))<Pw?this.object.lookAt(this.target):(mg.setFromNormalAndCoplanarPoint(this.object.up,this.target),Ol.intersectPlane(mg,this.target))))}else if(this.object.isOrthographicCamera){const u=this.object.zoom;this.object.zoom=Math.max(this.minZoom,Math.min(this.maxZoom,this.object.zoom/this._scale)),u!==this.object.zoom&&(this.object.updateProjectionMatrix(),l=!0)}return this._scale=1,this._performCursorZoom=!1,l||this._lastPosition.distanceToSquared(this.object.position)>yf||8*(1-this._lastQuaternion.dot(this.object.quaternion))>yf||this._lastTargetPosition.distanceToSquared(this.target)>yf?(this.dispatchEvent(pg),this._lastPosition.copy(this.object.position),this._lastQuaternion.copy(this.object.quaternion),this._lastTargetPosition.copy(this.target),!0):!1}_getAutoRotationAngle(e){return e!==null?Pn/60*this.autoRotateSpeed*e:Pn/60/60*this.autoRotateSpeed}_getZoomScale(e){const t=Math.abs(e*.01);return Math.pow(.95,this.zoomSpeed*t)}_rotateLeft(e){this._sphericalDelta.theta-=e}_rotateUp(e){this._sphericalDelta.phi-=e}_panLeft(e,t){Zt.setFromMatrixColumn(t,0),Zt.multiplyScalar(-e),this._panOffset.add(Zt)}_panUp(e,t){this.screenSpacePanning===!0?Zt.setFromMatrixColumn(t,1):(Zt.setFromMatrixColumn(t,0),Zt.crossVectors(this.object.up,Zt)),Zt.multiplyScalar(e),this._panOffset.add(Zt)}_pan(e,t){const s=this.domElement;if(this.object.isPerspectiveCamera){const a=this.object.position;Zt.copy(a).sub(this.target);let l=Zt.length();l*=Math.tan(this.object.fov/2*Math.PI/180),this._panLeft(2*e*l/s.clientHeight,this.object.matrix),this._panUp(2*t*l/s.clientHeight,this.object.matrix)}else this.object.isOrthographicCamera?(this._panLeft(e*(this.object.right-this.object.left)/this.object.zoom/s.clientWidth,this.object.matrix),this._panUp(t*(this.object.top-this.object.bottom)/this.object.zoom/s.clientHeight,this.object.matrix)):(console.warn("WARNING: OrbitControls.js encountered an unknown camera type - pan disabled."),this.enablePan=!1)}_dollyOut(e){this.object.isPerspectiveCamera||this.object.isOrthographicCamera?this._scale/=e:(console.warn("WARNING: OrbitControls.js encountered an unknown camera type - dolly/zoom disabled."),this.enableZoom=!1)}_dollyIn(e){this.object.isPerspectiveCamera||this.object.isOrthographicCamera?this._scale*=e:(console.warn("WARNING: OrbitControls.js encountered an unknown camera type - dolly/zoom disabled."),this.enableZoom=!1)}_updateZoomParameters(e,t){if(!this.zoomToCursor)return;this._performCursorZoom=!0;const s=this.domElement.getBoundingClientRect(),a=e-s.left,l=t-s.top,u=s.width,f=s.height;this._mouse.x=a/u*2-1,this._mouse.y=-(l/f)*2+1,this._dollyDirection.set(this._mouse.x,this._mouse.y,1).unproject(this.object).sub(this.object.position).normalize()}_clampDistance(e){return Math.max(this.minDistance,Math.min(this.maxDistance,e))}_handleMouseDownRotate(e){this._rotateStart.set(e.clientX,e.clientY)}_handleMouseDownDolly(e){this._updateZoomParameters(e.clientX,e.clientX),this._dollyStart.set(e.clientX,e.clientY)}_handleMouseDownPan(e){this._panStart.set(e.clientX,e.clientY)}_handleMouseMoveRotate(e){this._rotateEnd.set(e.clientX,e.clientY),this._rotateDelta.subVectors(this._rotateEnd,this._rotateStart).multiplyScalar(this.rotateSpeed);const t=this.domElement;this._rotateLeft(Pn*this._rotateDelta.x/t.clientHeight),this._rotateUp(Pn*this._rotateDelta.y/t.clientHeight),this._rotateStart.copy(this._rotateEnd),this.update()}_handleMouseMoveDolly(e){this._dollyEnd.set(e.clientX,e.clientY),this._dollyDelta.subVectors(this._dollyEnd,this._dollyStart),this._dollyDelta.y>0?this._dollyOut(this._getZoomScale(this._dollyDelta.y)):this._dollyDelta.y<0&&this._dollyIn(this._getZoomScale(this._dollyDelta.y)),this._dollyStart.copy(this._dollyEnd),this.update()}_handleMouseMovePan(e){this._panEnd.set(e.clientX,e.clientY),this._panDelta.subVectors(this._panEnd,this._panStart).multiplyScalar(this.panSpeed),this._pan(this._panDelta.x,this._panDelta.y),this._panStart.copy(this._panEnd),this.update()}_handleMouseWheel(e){this._updateZoomParameters(e.clientX,e.clientY),e.deltaY<0?this._dollyIn(this._getZoomScale(e.deltaY)):e.deltaY>0&&this._dollyOut(this._getZoomScale(e.deltaY)),this.update()}_handleKeyDown(e){let t=!1;switch(e.code){case this.keys.UP:e.ctrlKey||e.metaKey||e.shiftKey?this._rotateUp(Pn*this.rotateSpeed/this.domElement.clientHeight):this._pan(0,this.keyPanSpeed),t=!0;break;case this.keys.BOTTOM:e.ctrlKey||e.metaKey||e.shiftKey?this._rotateUp(-Pn*this.rotateSpeed/this.domElement.clientHeight):this._pan(0,-this.keyPanSpeed),t=!0;break;case this.keys.LEFT:e.ctrlKey||e.metaKey||e.shiftKey?this._rotateLeft(Pn*this.rotateSpeed/this.domElement.clientHeight):this._pan(this.keyPanSpeed,0),t=!0;break;case this.keys.RIGHT:e.ctrlKey||e.metaKey||e.shiftKey?this._rotateLeft(-Pn*this.rotateSpeed/this.domElement.clientHeight):this._pan(-this.keyPanSpeed,0),t=!0;break}t&&(e.preventDefault(),this.update())}_handleTouchStartRotate(e){if(this._pointers.length===1)this._rotateStart.set(e.pageX,e.pageY);else{const t=this._getSecondPointerPosition(e),s=.5*(e.pageX+t.x),a=.5*(e.pageY+t.y);this._rotateStart.set(s,a)}}_handleTouchStartPan(e){if(this._pointers.length===1)this._panStart.set(e.pageX,e.pageY);else{const t=this._getSecondPointerPosition(e),s=.5*(e.pageX+t.x),a=.5*(e.pageY+t.y);this._panStart.set(s,a)}}_handleTouchStartDolly(e){const t=this._getSecondPointerPosition(e),s=e.pageX-t.x,a=e.pageY-t.y,l=Math.sqrt(s*s+a*a);this._dollyStart.set(0,l)}_handleTouchStartDollyPan(e){this.enableZoom&&this._handleTouchStartDolly(e),this.enablePan&&this._handleTouchStartPan(e)}_handleTouchStartDollyRotate(e){this.enableZoom&&this._handleTouchStartDolly(e),this.enableRotate&&this._handleTouchStartRotate(e)}_handleTouchMoveRotate(e){if(this._pointers.length==1)this._rotateEnd.set(e.pageX,e.pageY);else{const s=this._getSecondPointerPosition(e),a=.5*(e.pageX+s.x),l=.5*(e.pageY+s.y);this._rotateEnd.set(a,l)}this._rotateDelta.subVectors(this._rotateEnd,this._rotateStart).multiplyScalar(this.rotateSpeed);const t=this.domElement;this._rotateLeft(Pn*this._rotateDelta.x/t.clientHeight),this._rotateUp(Pn*this._rotateDelta.y/t.clientHeight),this._rotateStart.copy(this._rotateEnd)}_handleTouchMovePan(e){if(this._pointers.length===1)this._panEnd.set(e.pageX,e.pageY);else{const t=this._getSecondPointerPosition(e),s=.5*(e.pageX+t.x),a=.5*(e.pageY+t.y);this._panEnd.set(s,a)}this._panDelta.subVectors(this._panEnd,this._panStart).multiplyScalar(this.panSpeed),this._pan(this._panDelta.x,this._panDelta.y),this._panStart.copy(this._panEnd)}_handleTouchMoveDolly(e){const t=this._getSecondPointerPosition(e),s=e.pageX-t.x,a=e.pageY-t.y,l=Math.sqrt(s*s+a*a);this._dollyEnd.set(0,l),this._dollyDelta.set(0,Math.pow(this._dollyEnd.y/this._dollyStart.y,this.zoomSpeed)),this._dollyOut(this._dollyDelta.y),this._dollyStart.copy(this._dollyEnd);const u=(e.pageX+t.x)*.5,f=(e.pageY+t.y)*.5;this._updateZoomParameters(u,f)}_handleTouchMoveDollyPan(e){this.enableZoom&&this._handleTouchMoveDolly(e),this.enablePan&&this._handleTouchMovePan(e)}_handleTouchMoveDollyRotate(e){this.enableZoom&&this._handleTouchMoveDolly(e),this.enableRotate&&this._handleTouchMoveRotate(e)}_addPointer(e){this._pointers.push(e.pointerId)}_removePointer(e){delete this._pointerPositions[e.pointerId];for(let t=0;t<this._pointers.length;t++)if(this._pointers[t]==e.pointerId){this._pointers.splice(t,1);return}}_isTrackingPointer(e){for(let t=0;t<this._pointers.length;t++)if(this._pointers[t]==e.pointerId)return!0;return!1}_trackPointer(e){let t=this._pointerPositions[e.pointerId];t===void 0&&(t=new rt,this._pointerPositions[e.pointerId]=t),t.set(e.pageX,e.pageY)}_getSecondPointerPosition(e){const t=e.pointerId===this._pointers[0]?this._pointers[1]:this._pointers[0];return this._pointerPositions[t]}_customWheelEvent(e){const t=e.deltaMode,s={clientX:e.clientX,clientY:e.clientY,deltaY:e.deltaY};switch(t){case 1:s.deltaY*=16;break;case 2:s.deltaY*=100;break}return e.ctrlKey&&!this._controlActive&&(s.deltaY*=10),s}}function Dw(r){this.enabled!==!1&&(this._pointers.length===0&&(this.domElement.setPointerCapture(r.pointerId),this.domElement.addEventListener("pointermove",this._onPointerMove),this.domElement.addEventListener("pointerup",this._onPointerUp)),!this._isTrackingPointer(r)&&(this._addPointer(r),r.pointerType==="touch"?this._onTouchStart(r):this._onMouseDown(r)))}function Nw(r){this.enabled!==!1&&(r.pointerType==="touch"?this._onTouchMove(r):this._onMouseMove(r))}function Iw(r){switch(this._removePointer(r),this._pointers.length){case 0:this.domElement.releasePointerCapture(r.pointerId),this.domElement.removeEventListener("pointermove",this._onPointerMove),this.domElement.removeEventListener("pointerup",this._onPointerUp),this.dispatchEvent(h_),this.state=Rt.NONE;break;case 1:const e=this._pointers[0],t=this._pointerPositions[e];this._onTouchStart({pointerId:e,pageX:t.x,pageY:t.y});break}}function Uw(r){let e;switch(r.button){case 0:e=this.mouseButtons.LEFT;break;case 1:e=this.mouseButtons.MIDDLE;break;case 2:e=this.mouseButtons.RIGHT;break;default:e=-1}switch(e){case Ks.DOLLY:if(this.enableZoom===!1)return;this._handleMouseDownDolly(r),this.state=Rt.DOLLY;break;case Ks.ROTATE:if(r.ctrlKey||r.metaKey||r.shiftKey){if(this.enablePan===!1)return;this._handleMouseDownPan(r),this.state=Rt.PAN}else{if(this.enableRotate===!1)return;this._handleMouseDownRotate(r),this.state=Rt.ROTATE}break;case Ks.PAN:if(r.ctrlKey||r.metaKey||r.shiftKey){if(this.enableRotate===!1)return;this._handleMouseDownRotate(r),this.state=Rt.ROTATE}else{if(this.enablePan===!1)return;this._handleMouseDownPan(r),this.state=Rt.PAN}break;default:this.state=Rt.NONE}this.state!==Rt.NONE&&this.dispatchEvent(Rd)}function Fw(r){switch(this.state){case Rt.ROTATE:if(this.enableRotate===!1)return;this._handleMouseMoveRotate(r);break;case Rt.DOLLY:if(this.enableZoom===!1)return;this._handleMouseMoveDolly(r);break;case Rt.PAN:if(this.enablePan===!1)return;this._handleMouseMovePan(r);break}}function Ow(r){this.enabled===!1||this.enableZoom===!1||this.state!==Rt.NONE||(r.preventDefault(),this.dispatchEvent(Rd),this._handleMouseWheel(this._customWheelEvent(r)),this.dispatchEvent(h_))}function kw(r){this.enabled===!1||this.enablePan===!1||this._handleKeyDown(r)}function Bw(r){switch(this._trackPointer(r),this._pointers.length){case 1:switch(this.touches.ONE){case qs.ROTATE:if(this.enableRotate===!1)return;this._handleTouchStartRotate(r),this.state=Rt.TOUCH_ROTATE;break;case qs.PAN:if(this.enablePan===!1)return;this._handleTouchStartPan(r),this.state=Rt.TOUCH_PAN;break;default:this.state=Rt.NONE}break;case 2:switch(this.touches.TWO){case qs.DOLLY_PAN:if(this.enableZoom===!1&&this.enablePan===!1)return;this._handleTouchStartDollyPan(r),this.state=Rt.TOUCH_DOLLY_PAN;break;case qs.DOLLY_ROTATE:if(this.enableZoom===!1&&this.enableRotate===!1)return;this._handleTouchStartDollyRotate(r),this.state=Rt.TOUCH_DOLLY_ROTATE;break;default:this.state=Rt.NONE}break;default:this.state=Rt.NONE}this.state!==Rt.NONE&&this.dispatchEvent(Rd)}function zw(r){switch(this._trackPointer(r),this.state){case Rt.TOUCH_ROTATE:if(this.enableRotate===!1)return;this._handleTouchMoveRotate(r),this.update();break;case Rt.TOUCH_PAN:if(this.enablePan===!1)return;this._handleTouchMovePan(r),this.update();break;case Rt.TOUCH_DOLLY_PAN:if(this.enableZoom===!1&&this.enablePan===!1)return;this._handleTouchMoveDollyPan(r),this.update();break;case Rt.TOUCH_DOLLY_ROTATE:if(this.enableZoom===!1&&this.enableRotate===!1)return;this._handleTouchMoveDollyRotate(r),this.update();break;default:this.state=Rt.NONE}}function Hw(r){this.enabled!==!1&&r.preventDefault()}function Vw(r){r.key==="Control"&&(this._controlActive=!0,this.domElement.getRootNode().addEventListener("keyup",this._interceptControlUp,{passive:!0,capture:!0}))}function Gw(r){r.key==="Control"&&(this._controlActive=!1,this.domElement.getRootNode().removeEventListener("keyup",this._interceptControlUp,{passive:!0,capture:!0}))}function Ww({trays:r,pending:e,fitsBox:t}){const s=he.useRef(null),[a,l]=he.useState(null);he.useEffect(()=>{if(!s.current)return;const f=s.current;l(null);const h=new vw;h.background=new vt(16249059);const p=new Zn(45,f.clientWidth/f.clientHeight,.1,4e3);p.position.set(180,180,240);const m=new _w({antialias:!0});m.setSize(f.clientWidth,f.clientHeight),f.innerHTML="",f.appendChild(m.domElement);const _=new Lw(p,m.domElement);_.enableDamping=!0,h.add(new Cw(16777215,.65));const x=new Aw(16777215,.8);x.position.set(1,1,1),h.add(x);const S=new bw,T=[];let w,v=!1;const y=t?8016432:11879215;async function P(){let D=0;for(const E of r){if(!E.url){D+=E.heightMm;continue}try{const C=await new Promise((ae,ue)=>{S.load(E.url,ae,void 0,ue)});if(v)return;C.computeVertexNormals(),C.center();const re=new xw({color:y}),ee=new xi(C,re);ee.position.y=D+E.heightMm/2,h.add(ee),T.push(ee)}catch{l("Could not load one of the tray previews.")}D+=E.heightMm}if(v||T.length===0)return;const V=new ao;for(const E of T)V.expandByObject(E);const O=new Q;V.getSize(O);const U=new Q;V.getCenter(U);const Y=Math.max(O.x,O.y,O.z,1);p.position.set(U.x+Y*1.1,U.y+Y*1.1,U.z+Y*1.4),_.target.copy(U);const ce=()=>{w=requestAnimationFrame(ce),_.update(),m.render(h,p)};ce()}P();const b=()=>{p.aspect=f.clientWidth/f.clientHeight,p.updateProjectionMatrix(),m.setSize(f.clientWidth,f.clientHeight)};return window.addEventListener("resize",b),()=>{v=!0,window.removeEventListener("resize",b),cancelAnimationFrame(w),T.forEach(D=>D.geometry.dispose()),m.dispose()}},[JSON.stringify(r),t]);const u=r.some(f=>f.url);return k.jsxs("div",{className:"card relative overflow-hidden bg-bgi-foam",style:{height:360},children:[k.jsx("div",{ref:s,className:"h-full w-full"}),e&&k.jsx("div",{className:"absolute inset-0 flex items-center justify-center bg-bgi-foam/80 text-sm text-bgi-ink/70",children:k.jsxs("span",{className:"flex items-center gap-2",children:[k.jsx("span",{className:"h-2 w-2 animate-shimmer rounded-full bg-bgi-teal"}),"Rendering preview…"]})}),!e&&!u&&k.jsx("div",{className:"absolute inset-0 flex items-center justify-center text-sm text-bgi-ink/50",children:"Adjust the parameters to see a preview."}),a&&k.jsx("div",{className:"absolute inset-0 flex items-center justify-center bg-bgi-foam/90 text-sm text-red-600",children:a})]})}function jw(r){const e=new URLSearchParams;for(const[t,s]of Object.entries(r))s==null||s===""||e.set(t,String(s));return e.toString()}function Xw(r,e){const t={};for(const[s,a]of Object.entries(r.properties)){const l=e.get(s);if(l===null)continue;const u=typeof a.type=="string"?a.type:a.type.find(f=>f&&f!=="null");if(u==="number"||u==="integer"){const f=Number(l);Number.isNaN(f)||(t[s]=f)}else u==="boolean"?t[s]=l==="true":t[s]=l}return t}const Yw={};function gg(r,e){const t={};for(const[s,a]of Object.entries(r.properties))a.default!==void 0?t[s]=a.default:a.minimum!==void 0?t[s]=a.minimum:a.enum&&a.enum.length>0&&(t[s]=a.enum[0]);if(t.boxProfileId===void 0){const s=e.find(a=>a.source==="original")??e[0];s&&(t.boxProfileId=s.id)}return t}function qw(){const{slug:r}=Jl(),[e,t]=sx(),s=md(),{addItem:a}=ec(),[l,u]=he.useState(null),[f,h]=he.useState(null),[p,m]=he.useState([]),[_,x]=he.useState({}),[S,T]=he.useState(null),[w,v]=he.useState(!1),[y,P]=he.useState(null),[b,D]=he.useState(!1),[V,O]=he.useState(null),[U,Y]=he.useState("idle"),ce=he.useMemo(()=>ta(),[]),E=he.useRef(0),C=he.useRef(null),re=he.useMemo(()=>{if(!r||!f)return;const F=Yw[r];if(!F)return;const se={},L={};for(const[X,ve]of Object.entries(F))ve.max&&(se[X]=ve.max(_,f)),ve.min&&(L[X]=ve.min(_,f));return{min:L,max:se}},[r,f,_]);he.useEffect(()=>{re&&x(F=>{let se=!1;const L={...F};for(const[X,ve]of Object.entries(re.max)){const Ne=Number(L[X]);Number.isFinite(Ne)&&Ne>ve&&(L[X]=ve,se=!0)}for(const[X,ve]of Object.entries(re.min)){const Ne=Number(L[X]);Number.isFinite(Ne)&&Ne<ve&&(L[X]=ve,se=!0)}return se?L:F})},[re]),he.useEffect(()=>{r&&Promise.all([Pt.getGame(r),Pt.getGameSchema(r),Pt.listSleeveProfiles()]).then(([F,se,L])=>{u(F),h(se),m(L);const X=Xw(se,e);x(Object.keys(X).length>0?{...gg(se,F.boxProfiles),...X}:gg(se,F.boxProfiles)),Pt.recordEvent("configurator_opened",{sessionId:ce,gameSlug:r}),Object.keys(X).length>0&&Pt.recordEvent("share_link_opened",{sessionId:ce,gameSlug:r})})},[r]),he.useEffect(()=>{if(!(!l?.productSlug||!f||Object.keys(_).length===0)&&!(!_.sleeveProfileId||!_.boxProfileId))return C.current&&clearTimeout(C.current),C.current=setTimeout(()=>{const F=++E.current;v(!0),P(null),Pt.preview(l.productSlug,_,ce).then(se=>{F===E.current&&(T(se),v(!1),Pt.recordEvent("preview_rendered",{sessionId:ce,gameSlug:r}),Pt.recordEvent("fit_indicator_shown",{sessionId:ce,gameSlug:r,metadata:{assembledHeightMm:se.assembledHeightMm,boxInteriorDepthMm:se.boxInteriorDepthMm,fits:se.fitsBox,boxVerified:se.boxVerified,depthIsPlaceholder:se.depthIsPlaceholder}}))}).catch(se=>{F===E.current&&(v(!1),P(se instanceof Lg?se.message:"Preview failed"))})},300),()=>{C.current&&clearTimeout(C.current)}},[l,f,_,ce,r]),he.useEffect(()=>{if(!V||V.status!=="pending")return;const F=setInterval(async()=>{const se=await Pt.getConfiguration(V.id);O(se),se.status!=="pending"&&(D(!1),clearInterval(F))},1500);return()=>clearInterval(F)},[V]);const ee=he.useCallback((F,se)=>{x(L=>({...L,[F]:se})),O(null),F==="sleeveProfileId"?Pt.recordEvent("sleeve_selected",{sessionId:ce,gameSlug:r}):F==="boxProfileId"?Pt.recordEvent("box_selected",{sessionId:ce,gameSlug:r}):Pt.recordEvent("parameter_changed",{sessionId:ce,gameSlug:r,metadata:{key:F}})},[ce,r]),ae=async()=>{if(!l?.productSlug)return;D(!0);const F=await Pt.validate(l.productSlug,_,ce),se=await Pt.getConfiguration(F.configurationId);O(se)};he.useEffect(()=>{V?.status==="valid"&&l?.productSlug&&(a({productSlug:l.productSlug,configurationId:V.id,quantity:1}),Pt.recordEvent("add_to_cart",{sessionId:ce,gameSlug:r,configurationId:V.id}),s("/cart")),V?.status==="rejected"&&Pt.recordEvent("fit_check_failed",{sessionId:ce,gameSlug:r,configurationId:V.id,rule:"set_validation",metadata:{reason:V.rejectionReason}})},[V,l,r,ce,a,s]);const ue=()=>{const F=jw(_),se=`${window.location.origin}/configure/${r}?${F}`;t(F),navigator.clipboard?.writeText(se).then(()=>{Y("copied"),Pt.recordEvent("share_link_created",{sessionId:ce,gameSlug:r}),setTimeout(()=>Y("idle"),2e3)})};if(!l||!f)return k.jsx("p",{className:"text-bgi-ink/60",children:"Loading…"});const Z=V?.status==="valid"&&V.trays?V.trays.map(F=>({url:F.stlUrl??"",heightMm:F.heightMm})):S?.firstTrayPreviewUrl?[{url:S.firstTrayPreviewUrl,heightMm:S.assembledHeightMm}]:[],le=S&&(!S.boxVerified||S.depthIsPlaceholder);return k.jsxs("div",{className:"grid gap-8 md:grid-cols-2",children:[k.jsxs("div",{children:[k.jsxs("h1",{className:"font-display mb-4 text-2xl font-bold text-bgi-ink",children:[l.name," Tray Set"]}),k.jsx(Ww,{trays:Z,pending:w,fitsBox:S?.fitsBox??!0}),S&&k.jsxs("div",{className:"mt-3 space-y-1 text-sm",children:[k.jsxs("p",{className:"text-bgi-ink/70",children:["Assembled height: ",S.assembledHeightMm.toFixed(0),"mm of ",S.boxInteriorDepthMm.toFixed(0),"mm box depth",S.trayCount>1?` (${S.trayCount} trays)`:""]}),k.jsx("p",{className:S.fitsBox?"font-medium text-bgi-teal":"font-medium text-red-600",children:S.fitsBox?"✓ Fits the box, lid flush":"✗ Won't fit — try a thinner sleeve class or a deeper box"}),S.unassembledComponents&&S.unassembledComponents.length>0&&k.jsxs("p",{className:"text-bgi-coral",children:["No tray design yet for: ",S.unassembledComponents.join(", ")," — these won't be included in this set."]})]}),y&&k.jsx("p",{className:"mt-2 text-sm text-red-600",children:y}),le&&k.jsxs("p",{className:"banner-caution mt-3",children:["Box and/or sleeve dimensions for this game are sourced from published references and not yet physically verified. See"," ",k.jsx(yi,{to:`/games/${r}/compatibility`,className:"underline decoration-bgi-coral/50 underline-offset-2",children:"compatibility notes"}),"."]}),k.jsx("button",{onClick:ue,className:"mt-4 text-sm font-medium text-bgi-teal underline decoration-bgi-teal/40 underline-offset-2 hover:text-bgi-coral",children:U==="copied"?"Link copied!":"Copy shareable link"})]}),k.jsxs("div",{className:"card p-6",children:[k.jsx(fx,{schema:f,values:_,onChange:ee,sleeveProfiles:p,boxProfiles:l.boxProfiles,derivedMax:re?.max,derivedMin:re?.min}),k.jsxs("div",{className:"mt-6",children:[V?.status==="rejected"&&k.jsx("p",{className:"mb-3 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700",children:V.rejectionReason}),k.jsx("button",{onClick:ae,disabled:b,className:"btn-primary w-full",children:b?"Validating…":"Add to cart"}),k.jsx("p",{className:"mt-2 text-xs text-bgi-ink/50",children:"Adding to cart generates and slices every tray in the set — this can take a few minutes."})]})]})]})}function $w(){const{items:r,removeItem:e,setQuantity:t}=ec(),[s,a]=he.useState(null),[l,u]=he.useState(""),[f,h]=he.useState(null),[p,m]=he.useState(!1);he.useEffect(()=>{r.length!==0&&Pt.cart(r).then(a)},[r]);const _=async()=>{if(!l){h("Enter an email address to continue.");return}m(!0),h(null),Pt.recordEvent("checkout_started",{sessionId:ta()});try{const x=await Pt.checkout(r,l,`${window.location.origin}/orders/:orderToken:`,`${window.location.origin}/cart`,ta());window.location.href=x.checkoutUrl}catch(x){h(x instanceof Error?x.message:"Checkout failed"),m(!1)}};return r.length===0?k.jsxs("div",{className:"card mx-auto max-w-md space-y-3 p-8 text-center",children:[k.jsx("p",{className:"text-bgi-ink/70",children:"Your cart is empty."}),k.jsx(yi,{to:"/",className:"btn-secondary",children:"Back to the catalog"})]}):k.jsxs("div",{className:"mx-auto max-w-lg space-y-6",children:[k.jsx("h1",{className:"font-display text-2xl font-bold text-bgi-ink",children:"Cart"}),!s&&k.jsx("p",{className:"text-bgi-ink/60",children:"Loading…"}),s&&k.jsxs(k.Fragment,{children:[k.jsx("ul",{className:"card divide-y divide-bgi-teal/10 px-5",children:s.items.map((x,S)=>k.jsxs("li",{className:"flex items-center justify-between py-4",children:[k.jsxs("div",{children:[k.jsx("p",{className:"font-medium text-bgi-ink",children:x.productName}),k.jsxs("p",{className:"text-sm text-bgi-ink/60",children:["$",(x.unitPriceCents/100).toFixed(2)," each"]}),k.jsxs("div",{className:"mt-2 flex items-center gap-2 text-sm",children:[k.jsx("button",{className:"btn-ghost","aria-label":"Decrease quantity",onClick:()=>t(r[S],Math.max(0,x.quantity-1)),children:"−"}),k.jsx("span",{className:"w-4 text-center tabular-nums",children:x.quantity}),k.jsx("button",{className:"btn-ghost","aria-label":"Increase quantity",onClick:()=>t(r[S],x.quantity+1),children:"+"}),k.jsx("button",{className:"ml-2 text-xs font-medium text-red-500 underline underline-offset-2 hover:text-red-600",onClick:()=>e(r[S]),children:"Remove"})]})]}),k.jsxs("p",{className:"font-medium text-bgi-ink",children:["$",(x.lineTotalCents/100).toFixed(2)]})]},S))}),k.jsxs("div",{className:"card space-y-1 p-5 text-sm",children:[k.jsxs("div",{className:"flex justify-between",children:[k.jsx("span",{className:"text-bgi-ink/70",children:"Subtotal"}),k.jsxs("span",{children:["$",(s.subtotalCents/100).toFixed(2)]})]}),k.jsxs("div",{className:"flex justify-between",children:[k.jsx("span",{className:"text-bgi-ink/70",children:"Shipping"}),k.jsx("span",{children:s.shippingCents===0?"Free":`$${(s.shippingCents/100).toFixed(2)}`})]}),k.jsxs("div",{className:"flex justify-between border-t border-bgi-teal/10 pt-2 text-base font-semibold text-bgi-ink",children:[k.jsx("span",{children:"Total"}),k.jsxs("span",{children:["$",(s.totalCents/100).toFixed(2)]})]})]}),k.jsxs("div",{children:[k.jsx("label",{className:"mb-1 block text-sm font-medium text-bgi-ink",children:"Email"}),k.jsx("input",{type:"email",className:"input-field",value:l,onChange:x=>u(x.target.value),placeholder:"you@example.com"})]}),k.jsx("p",{className:"text-xs text-bgi-ink/50",children:"This is a made-to-order print — expect several days of fulfillment lead time before it ships."}),f&&k.jsx("p",{className:"text-sm text-red-600",children:f}),k.jsx("button",{onClick:_,disabled:p,className:"btn-primary w-full",children:p?"Redirecting to checkout…":"Checkout"})]})]})}function Kw(){const{token:r}=Jl(),[e,t]=he.useState(void 0),{clear:s}=ec();return he.useEffect(()=>{r&&Pt.getOrder(r).then(a=>{t(a),s()}).catch(()=>t(null))},[r]),e===void 0?k.jsx("p",{className:"text-bgi-ink/60",children:"Loading…"}):e===null?k.jsx("p",{className:"text-bgi-ink/60",children:"We couldn't find that order."}):k.jsxs("div",{className:"mx-auto max-w-lg space-y-5",children:[k.jsxs("h1",{className:"font-display text-2xl font-bold text-bgi-ink",children:["Order ",e.orderToken]}),k.jsxs("p",{className:"pill bg-bgi-teal/10 text-bgi-teal",children:["Status: ",k.jsx("span",{className:"ml-1 font-semibold",children:Zw(e.status)})]}),k.jsx("ul",{className:"card divide-y divide-bgi-teal/10 px-5",children:e.items.map(a=>k.jsxs("li",{className:"flex justify-between py-3 text-sm",children:[k.jsxs("span",{className:"text-bgi-ink/80",children:[a.productName||"Tray set"," × ",a.quantity]}),k.jsxs("span",{className:"font-medium text-bgi-ink",children:["$",(a.unitPriceCents*a.quantity/100).toFixed(2)]})]},a.id))}),k.jsxs("div",{className:"card space-y-1 p-5 text-sm",children:[k.jsxs("div",{className:"flex justify-between",children:[k.jsx("span",{className:"text-bgi-ink/70",children:"Subtotal"}),k.jsxs("span",{children:["$",(e.subtotalCents/100).toFixed(2)]})]}),k.jsxs("div",{className:"flex justify-between",children:[k.jsx("span",{className:"text-bgi-ink/70",children:"Shipping"}),k.jsxs("span",{children:["$",(e.shippingCents/100).toFixed(2)]})]}),k.jsxs("div",{className:"flex justify-between border-t border-bgi-teal/10 pt-2 font-semibold text-bgi-ink",children:[k.jsx("span",{children:"Total"}),k.jsxs("span",{children:["$",(e.totalCents/100).toFixed(2)]})]})]}),e.status==="pending_payment"&&k.jsx("p",{className:"text-xs text-bgi-ink/50",children:"Once payment clears, this order is generated as N supportless, single-plate trays and queued for printing — expect several days of fulfillment lead time."})]})}function Zw(r){switch(r){case"pending_payment":return"Awaiting payment";case"paid":return"Paid — queued for printing";case"fulfilled":return"Shipped";case"cancelled":return"Cancelled";default:return r}}function Qw(){const{slug:r}=Jl(),[e,t]=he.useState(null),[s,a]=he.useState(null);return he.useEffect(()=>{r&&(Pt.getGame(r).then(t),Pt.getGameCompatibility(r).then(a))},[r]),k.jsxs("div",{className:"mx-auto max-w-2xl space-y-6",children:[k.jsx("h1",{className:"font-display text-2xl font-bold text-bgi-ink",children:"Compatibility & disclaimer"}),k.jsxs("div",{className:"card space-y-3 p-6 text-sm text-bgi-ink/80",children:[k.jsxs("p",{children:["Products on this site are organizer trays sized to fit specific commercial board games, described using the game's name for the sole purpose of indicating compatibility (nominative reference). This site is"," ",k.jsx("strong",{children:"not affiliated with, endorsed by, or sponsored by"})," the publisher of ",e?.name??"any game listed here",e?.publisher?` (${e.publisher})`:""," or any other game referenced on this site."]}),k.jsx("p",{children:"No box art, logos, trademarked graphics, or fonts belonging to any publisher are reproduced on this site or on any printed part — only functional geometry sized to publicly known, factual component dimensions and counts."}),k.jsx("p",{children:"Box dimensions can vary between printings and regional editions. Component counts and card sizes are sourced from published community references and may be revised. No warranty of exact fit is made for any game not yet marked verified below."})]}),e&&s&&k.jsxs("div",{className:"card space-y-4 p-6",children:[k.jsxs("h2",{className:"font-display text-lg font-bold text-bgi-ink",children:[e.name," — data sources"]}),k.jsxs("div",{children:[k.jsx("h3",{className:"mb-1 text-sm font-semibold text-bgi-ink",children:"Box dimensions"}),k.jsx("ul",{className:"space-y-1 text-sm text-bgi-ink/70",children:s.boxProfiles.map(l=>k.jsxs("li",{children:[l.label," (",l.source,"): ",l.interiorLengthMm,"×",l.interiorWidthMm,"×",l.interiorDepthMm,"mm interior —"," ",l.verified?k.jsx("span",{className:"text-bgi-teal",children:"verified"}):k.jsxs("span",{className:"text-bgi-coral",children:["unverified",l.depthIsPlaceholder?", depth is a placeholder":""]}),l.measurementNotes&&k.jsx("span",{className:"block text-xs italic",children:l.measurementNotes})]},l.id))})]}),k.jsxs("div",{children:[k.jsx("h3",{className:"mb-1 text-sm font-semibold text-bgi-ink",children:"Component counts"}),k.jsx("ul",{className:"space-y-1 text-sm text-bgi-ink/70",children:s.manifest.map(l=>k.jsxs("li",{children:[l.componentType,": ",l.count," —"," ",l.verified?k.jsx("span",{className:"text-bgi-teal",children:"verified"}):k.jsx("span",{className:"text-bgi-coral",children:"unverified"}),l.notes&&k.jsx("span",{className:"block text-xs italic",children:l.notes})]},l.id))})]})]}),k.jsx(yi,{to:"/",className:"btn-secondary",children:"Back to the catalog"})]})}function Jw(){return k.jsxs(yi,{to:"/",className:"group flex items-center gap-2",children:[k.jsx("span",{className:"flex h-9 w-9 items-center justify-center rounded-full border-2 border-bgi-ink bg-bgi-glow shadow-pop-sm transition-all duration-300 ease-bounce group-hover:-translate-y-0.5 group-hover:rotate-12",children:k.jsxs("svg",{viewBox:"0 0 24 24",className:"h-4 w-4 text-bgi-ink","aria-hidden":"true",children:[k.jsx("rect",{x:"4",y:"4",width:"16",height:"16",rx:"3",fill:"none",stroke:"currentColor",strokeWidth:"2"}),k.jsx("circle",{cx:"9",cy:"9",r:"1.3",fill:"currentColor"}),k.jsx("circle",{cx:"15",cy:"15",r:"1.3",fill:"currentColor"}),k.jsx("circle",{cx:"9",cy:"15",r:"1.3",fill:"currentColor",opacity:"0.6"}),k.jsx("circle",{cx:"15",cy:"9",r:"1.3",fill:"currentColor",opacity:"0.6"})]})}),k.jsx("span",{className:"font-display text-xl font-bold tracking-tight text-bgi-ink",children:"trays"})]})}function e1({children:r}){const{items:e}=ec(),t=e.reduce((s,a)=>s+a.quantity,0);return k.jsxs("div",{className:"flex min-h-screen flex-col bg-bgi-paper",children:[k.jsx("header",{className:"sticky top-0 z-20 border-b-2 border-bgi-ink/10 bg-bgi-paper/90 backdrop-blur",children:k.jsxs("nav",{className:"mx-auto flex max-w-5xl items-center justify-between px-4 py-3",children:[k.jsx(Jw,{}),k.jsx("div",{className:"flex items-center gap-3 text-xs font-medium sm:gap-6 sm:text-sm",children:k.jsxs(yi,{to:"/cart",className:"flex items-center gap-1.5 rounded-full border-2 border-bgi-ink bg-white px-3 py-1.5 font-bold shadow-pop-sm transition-all duration-300 ease-bounce hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[5px_5px_0_0_#241a12] active:translate-x-0.5 active:translate-y-0.5 active:shadow-none",children:["Cart",t>0&&k.jsx("span",{className:"flex h-5 min-w-5 animate-pop-in items-center justify-center rounded-full bg-bgi-coral px-1 text-xs font-bold text-white",children:t})]})})]})}),k.jsx("main",{className:"mx-auto w-full max-w-5xl flex-1 px-4 py-10",children:r}),k.jsx("footer",{className:"mt-8 bg-bgi-lagoon text-bgi-sand/70",children:k.jsxs("div",{className:"mx-auto max-w-5xl px-4 py-8 text-xs",children:[k.jsxs("div",{className:"mb-3 flex items-center gap-2 font-display text-sm text-bgi-sand",children:[k.jsx("span",{className:"h-1.5 w-1.5 rounded-full bg-bgi-glow"}),"trays"]}),"PETG parts, made to order. Not affiliated with any game publisher — see"," ",k.jsx(yi,{to:"/games",className:"underline decoration-bgi-glow/50 underline-offset-2 hover:text-bgi-glow",children:"compatibility notes"}),"."]})})]})}function t1(){return k.jsx(tx,{children:k.jsx(e1,{children:k.jsxs(Y0,{children:[k.jsx(Kr,{path:"/",element:k.jsx(ax,{})}),k.jsx(Kr,{path:"/games/:slug",element:k.jsx(ux,{})}),k.jsx(Kr,{path:"/games/:slug/compatibility",element:k.jsx(Qw,{})}),k.jsx(Kr,{path:"/configure/:slug",element:k.jsx(qw,{})}),k.jsx(Kr,{path:"/cart",element:k.jsx($w,{})}),k.jsx(Kr,{path:"/orders/:token",element:k.jsx(Kw,{})})]})})})}r0.createRoot(document.getElementById("root")).render(k.jsx(he.StrictMode,{children:k.jsx(t1,{})}));
