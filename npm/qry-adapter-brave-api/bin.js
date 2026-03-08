#!/usr/bin/env node
"use strict";
const { spawnSync } = require("child_process");
const path = require("path");
const fs = require("fs");

const exe = process.platform === "win32" ? "qry-adapter-brave-api.exe" : "qry-adapter-brave-api";
const binPath = path.join(__dirname, "bin", exe);

if (!fs.existsSync(binPath)) {
  console.error(`qry-adapter-brave-api: binary not found at ${binPath}`);
  console.error("Try reinstalling: npm install @justestif/qry-adapter-brave-api");
  process.exit(1);
}

const result = spawnSync(binPath, process.argv.slice(2), { stdio: "inherit" });
process.exit(result.status ?? 1);
