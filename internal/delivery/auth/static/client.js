(() => {
  "use strict";

  const accessTokenKey = "zitadel_access_token";
  const idTokenKey = "zitadel_id_token";
  const stateKey = "zitadel_oauth_state";
  const nonceKey = "zitadel_oauth_nonce";
  const verifierKey = "zitadel_pkce_verifier";
  const returnToKey = "zitadel_return_to";

  function accessToken() {
    return sessionStorage.getItem(accessTokenKey);
  }

  function clearTokens() {
    sessionStorage.removeItem(accessTokenKey);
    sessionStorage.removeItem(idTokenKey);
  }

  function login(returnTo = window.location.pathname + window.location.search) {
    window.location.assign("/auth/login?return_to=" + encodeURIComponent(safeReturnTo(returnTo)));
  }

  async function authenticatedFetch(input, init = {}) {
    const token = accessToken();
    if (!token) {
      login();
      throw new Error("authentication required");
    }

    const headers = new Headers(init.headers || {});
    headers.set("Authorization", "Bearer " + token);
    const response = await fetch(input, {...init, headers});
    if (response.status === 401) {
      clearTokens();
      login();
      throw new Error("authentication expired");
    }
    return response;
  }

  function safeReturnTo(value) {
    try {
      const target = new URL(value || "/dashboard", window.location.origin);
      if (target.origin === window.location.origin && target.pathname.startsWith("/dashboard")) {
        return target.pathname + target.search + target.hash;
      }
    } catch (_) {
    }
    return "/dashboard";
  }

  function randomBase64URL(byteLength) {
    const bytes = crypto.getRandomValues(new Uint8Array(byteLength));
    let binary = "";
    for (const byte of bytes) {
      binary += String.fromCharCode(byte);
    }
    return btoa(binary).replaceAll("+", "-").replaceAll("/", "_").replaceAll("=", "");
  }

  async function codeChallenge(verifier) {
    const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(verifier));
    let binary = "";
    for (const byte of new Uint8Array(digest)) {
      binary += String.fromCharCode(byte);
    }
    return btoa(binary).replaceAll("+", "-").replaceAll("/", "_").replaceAll("=", "");
  }

  async function loadConfig() {
    const response = await fetch("/auth/config", {cache: "no-store", credentials: "omit"});
    if (!response.ok) {
      throw new Error("authentication configuration is unavailable");
    }
    return response.json();
  }

  function tokenClaims(token) {
    const encoded = token.split(".")[1];
    if (!encoded) {
      throw new Error("identity token is malformed");
    }
    const normalized = encoded.replaceAll("-", "+").replaceAll("_", "/");
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=");
    const bytes = Uint8Array.from(atob(padded), character => character.charCodeAt(0));
    return JSON.parse(new TextDecoder().decode(bytes));
  }

  function setStatus(message, isError = false) {
    const status = document.getElementById("auth-status");
    if (!status) {
      return;
    }
    status.textContent = message;
    status.classList.toggle("error", isError);
  }

  async function startLogin() {
    const config = await loadConfig();
    const verifier = randomBase64URL(64);
    const state = randomBase64URL(32);
    const nonce = randomBase64URL(32);
    const returnTo = safeReturnTo(new URLSearchParams(window.location.search).get("return_to"));

    sessionStorage.setItem(verifierKey, verifier);
    sessionStorage.setItem(stateKey, state);
    sessionStorage.setItem(nonceKey, nonce);
    sessionStorage.setItem(returnToKey, returnTo);

    const parameters = new URLSearchParams({
      client_id: config.clientId,
      redirect_uri: config.redirectUri,
      response_type: "code",
      scope: config.scopes.join(" "),
      state,
      nonce,
      code_challenge: await codeChallenge(verifier),
      code_challenge_method: "S256"
    });
    window.location.replace(config.authorizationEndpoint + "?" + parameters);
  }

  async function finishLogin() {
    const parameters = new URLSearchParams(window.location.search);
    const expectedState = sessionStorage.getItem(stateKey);
    const expectedNonce = sessionStorage.getItem(nonceKey);
    const verifier = sessionStorage.getItem(verifierKey);
    const code = parameters.get("code");

    if (parameters.get("error")) {
      throw new Error(parameters.get("error_description") || parameters.get("error"));
    }
    if (!code || !expectedState || parameters.get("state") !== expectedState || !verifier) {
      throw new Error("authentication response validation failed");
    }

    const config = await loadConfig();
    const body = new URLSearchParams({
      grant_type: "authorization_code",
      code,
      redirect_uri: config.redirectUri,
      client_id: config.clientId,
      code_verifier: verifier
    });
    const response = await fetch(config.tokenEndpoint, {
      method: "POST",
      headers: {"Content-Type": "application/x-www-form-urlencoded"},
      body,
      credentials: "omit"
    });
    if (!response.ok) {
      throw new Error("token exchange failed");
    }

    const tokens = await response.json();
    if (!tokens.access_token || !tokens.id_token || tokenClaims(tokens.id_token).nonce !== expectedNonce) {
      throw new Error("identity token validation failed");
    }

    sessionStorage.setItem(accessTokenKey, tokens.access_token);
    sessionStorage.setItem(idTokenKey, tokens.id_token);
    const returnTo = safeReturnTo(sessionStorage.getItem(returnToKey));
    sessionStorage.removeItem(verifierKey);
    sessionStorage.removeItem(stateKey);
    sessionStorage.removeItem(nonceKey);
    sessionStorage.removeItem(returnToKey);
    window.location.replace(returnTo);
  }

  async function logout() {
    const idToken = sessionStorage.getItem(idTokenKey);
    const config = await loadConfig();
    clearTokens();
    const parameters = new URLSearchParams({
      client_id: config.clientId,
      post_logout_redirect_uri: config.postLogoutRedirectUri
    });
    if (idToken) {
      parameters.set("id_token_hint", idToken);
    }
    window.location.replace(config.endSessionEndpoint + "?" + parameters);
  }

  document.addEventListener("htmx:config:request", event => {
    const token = accessToken();
    if (token) {
      event.detail.ctx.request.headers.Authorization = "Bearer " + token;
    }
  });

  document.addEventListener("htmx:response:error", event => {
    if (event.detail.ctx.response.status === 401) {
      clearTokens();
      login();
    }
  });

  window.zitadelAuth = Object.freeze({fetch: authenticatedFetch, login});

  const flow = document.body?.dataset.authFlow;
  const actions = {login: startLogin, callback: finishLogin, logout};
  if (actions[flow]) {
    actions[flow]().catch(error => setStatus(error.message, true));
  }
})();