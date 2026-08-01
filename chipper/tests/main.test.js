const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const vm = require("node:vm");

const mainJsPath = path.join(__dirname, "..", "webroot", "js", "main.js");
const mainJsSource = fs.readFileSync(mainJsPath, "utf8");

function extractFunction(source, name) {
  const start = source.indexOf(`function ${name}(`);
  if (start === -1) {
    throw new Error(`could not find function ${name}`);
  }

  const openBrace = source.indexOf("{", start);
  let depth = 0;
  for (let i = openBrace; i < source.length; i += 1) {
    const char = source[i];
    if (char === "{") {
      depth += 1;
    } else if (char === "}") {
      depth -= 1;
      if (depth === 0) {
        return source.slice(start, i + 1);
      }
    }
  }

  throw new Error(`could not extract function ${name}`);
}

function createElement(id) {
  return {
    id,
    style: { display: "" },
    value: "",
    checked: false,
    innerHTML: "",
    textContent: "",
    children: [],
    appendChild(child) {
      this.children.push(child);
      return child;
    },
    addEventListener() {},
    classList: {
      add() {},
      remove() {},
    },
  };
}

function createSandbox(fetchImpl) {
  const elements = new Map();
  const alerts = [];

  const document = {
    getElementById(id) {
      if (!elements.has(id)) {
        elements.set(id, createElement(id));
      }
      return elements.get(id);
    },
    createElement(tagName) {
      return createElement(tagName);
    },
  };

  const sandbox = {
    document,
    fetch: fetchImpl,
    alert(message) {
      alerts.push(message);
    },
    console,
    JSON,
    Promise,
  };

  sandbox.getE = (id) => document.getElementById(id);

  vm.createContext(sandbox);
  vm.runInContext(
    [
      extractFunction(mainJsSource, "displayMessage"),
      extractFunction(mainJsSource, "checkKG"),
      extractFunction(mainJsSource, "sendKGAPIKey"),
      extractFunction(mainJsSource, "updateKGAPI"),
    ].join("\n\n"),
    sandbox
  );

  return {
    sandbox,
    alerts,
    getElement(id) {
      return document.getElementById(id);
    },
  };
}

async function flushPromises() {
  await new Promise((resolve) => setImmediate(resolve));
  await new Promise((resolve) => setImmediate(resolve));
}

test("checkKG toggles DeepSeek and custom sections", () => {
  const { sandbox, getElement } = createSandbox(() => Promise.reject(new Error("unused fetch")));

  getElement("kgProvider").value = "deepseek";
  sandbox.checkKG();
  assert.equal(getElement("deepSeekInput").style.display, "block");
  assert.equal(getElement("intentGraphInput").style.display, "block");
  assert.equal(getElement("saveChatInput").style.display, "block");
  assert.equal(getElement("llmCommandInput").style.display, "block");
  assert.equal(getElement("customAIInput").style.display, "none");
  assert.equal(getElement("openAIInput").style.display, "none");

  getElement("kgProvider").value = "custom";
  sandbox.checkKG();
  assert.equal(getElement("customAIInput").style.display, "block");
  assert.equal(getElement("deepSeekInput").style.display, "none");
  assert.equal(getElement("openAIInput").style.display, "none");
});

test("sendKGAPIKey saves DeepSeek payload with default model", async () => {
  let fetchCall;
  const { sandbox, alerts, getElement } = createSandbox((url, init) => {
    fetchCall = { url, init };
    return Promise.resolve({
      text: () => Promise.resolve("saved"),
    });
  });

  getElement("kgProvider").value = "deepseek";
  getElement("deepseekKey").value = "deepseek-key";
  getElement("deepseekModel").value = "";
  getElement("deepseekPrompt").value = "Speak simply";
  getElement("intentyes").checked = true;
  getElement("saveChatYes").checked = true;
  getElement("commandYes").checked = false;

  sandbox.sendKGAPIKey();
  await flushPromises();

  assert.equal(fetchCall.url, "/api/set_kg_api");
  assert.equal(fetchCall.init.method, "POST");
  assert.equal(fetchCall.init.headers["Content-Type"], "application/json");
  assert.deepEqual(JSON.parse(fetchCall.init.body), {
    enable: true,
    provider: "deepseek",
    key: "deepseek-key",
    model: "deepseek-v4-flash",
    id: "",
    intentgraph: true,
    robotName: "",
    openai_prompt: "Speak simply",
    openai_voice: "",
    openai_voice_with_english: false,
    save_chat: true,
    commands_enable: false,
    endpoint: "",
  });
  assert.deepEqual(alerts, ["saved"]);
});

test("sendKGAPIKey saves custom payload with endpoint", async () => {
  let fetchCall;
  const { sandbox, alerts, getElement } = createSandbox((url, init) => {
    fetchCall = { url, init };
    return Promise.resolve({
      text: () => Promise.resolve("custom saved"),
    });
  });

  getElement("kgProvider").value = "custom";
  getElement("customKey").value = "ollama";
  getElement("customModel").value = "qwen:7b";
  getElement("customAIPrompt").value = "Stay practical";
  getElement("customAIEndpoint").value = "http://127.0.0.1:11434/v1";
  getElement("intentyes").checked = true;
  getElement("saveChatYes").checked = false;
  getElement("commandYes").checked = true;

  sandbox.sendKGAPIKey();
  await flushPromises();

  assert.deepEqual(JSON.parse(fetchCall.init.body), {
    enable: true,
    provider: "custom",
    key: "ollama",
    model: "qwen:7b",
    id: "",
    intentgraph: true,
    robotName: "",
    openai_prompt: "Stay practical",
    openai_voice: "",
    openai_voice_with_english: false,
    save_chat: false,
    commands_enable: true,
    endpoint: "http://127.0.0.1:11434/v1",
  });
  assert.deepEqual(alerts, ["custom saved"]);
});

test("updateKGAPI refills DeepSeek fields and applies default model", async () => {
  const { sandbox, getElement } = createSandbox((url) => {
    assert.equal(url, "/api/get_kg_api");
    return Promise.resolve({
      json: () =>
        Promise.resolve({
          provider: "deepseek",
          key: "deepseek-key",
          model: "",
          openai_prompt: "Refilled prompt",
          commands_enable: true,
          intentgraph: false,
          save_chat: true,
        }),
    });
  });

  sandbox.updateKGAPI();
  await flushPromises();

  assert.equal(getElement("kgProvider").value, "deepseek");
  assert.equal(getElement("deepseekKey").value, "deepseek-key");
  assert.equal(getElement("deepseekModel").value, "deepseek-v4-flash");
  assert.equal(getElement("deepseekPrompt").value, "Refilled prompt");
  assert.equal(getElement("commandYes").checked, true);
  assert.equal(getElement("intentyes").checked, false);
  assert.equal(getElement("saveChatYes").checked, true);
  assert.equal(getElement("deepSeekInput").style.display, "block");
  assert.equal(getElement("customAIInput").style.display, "none");
});

test("updateKGAPI refills custom fields including endpoint", async () => {
  const { sandbox, getElement } = createSandbox((url) => {
    assert.equal(url, "/api/get_kg_api");
    return Promise.resolve({
      json: () =>
        Promise.resolve({
          provider: "custom",
          key: "ollama",
          model: "qwen:7b",
          openai_prompt: "Custom prompt",
          endpoint: "http://127.0.0.1:11434/v1",
          commands_enable: false,
          intentgraph: true,
          save_chat: false,
        }),
    });
  });

  sandbox.updateKGAPI();
  await flushPromises();

  assert.equal(getElement("kgProvider").value, "custom");
  assert.equal(getElement("customKey").value, "ollama");
  assert.equal(getElement("customModel").value, "qwen:7b");
  assert.equal(getElement("customAIPrompt").value, "Custom prompt");
  assert.equal(getElement("customAIEndpoint").value, "http://127.0.0.1:11434/v1");
  assert.equal(getElement("commandYes").checked, false);
  assert.equal(getElement("intentyes").checked, true);
  assert.equal(getElement("saveChatYes").checked, false);
  assert.equal(getElement("customAIInput").style.display, "block");
  assert.equal(getElement("deepSeekInput").style.display, "none");
});
