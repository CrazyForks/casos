import {message} from "antd";
import Sdk from "casdoor-js-sdk";

export let ServerUrl = "";

export let CasdoorSdk;

export function initServerUrl() {
  const fullServerUrl = window.location.origin;
  if (fullServerUrl === "http://localhost:8001") {
    ServerUrl = "http://localhost:9000";
  }
}

export function initCasdoorSdk(config) {
  CasdoorSdk = new Sdk(config);
}

export function getSigninUrl() {
  return CasdoorSdk.getSigninUrl();
}

export function getSignupUrl() {
  return CasdoorSdk.getSignupUrl();
}

export function getUserProfileUrl(userName, account) {
  return CasdoorSdk.getUserProfileUrl(userName, account);
}

export function getMyProfileUrl(account) {
  return CasdoorSdk.getMyProfileUrl(account);
}

export function signin() {
  return CasdoorSdk.signin(ServerUrl);
}

export function goToLink(link) {
  window.location.href = link;
}

export function showMessage(type, msg) {
  if (type === "success") {
    message.success(msg);
  } else if (type === "error") {
    message.error(msg);
  } else if (type === "info") {
    message.info(msg);
  }
}

export function getItem(label, key, icon, children) {
  return {key, icon, children, label};
}

export function getAcceptLanguage() {
  return "en";
}

export function isMobile() {
  return window.innerWidth < 768;
}
