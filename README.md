# 🛡️ ArchGuard

> **An Open-Source Engineering Policy Engine (ESLint for Engineering Policies)**
> 
> *เปลี่ยนกฎการเขียนโค้ด มาตรฐานความปลอดภัย และสถาปัตยกรรมขององค์กร ให้กลายเป็นการตรวจสอบอัตโนมัติ (Automated Engineering Policy Engine)*

[![Release](https://img.shields.io/badge/Release-v0.4.1-blue.svg)](https://github.com/jakkayy/archGuard/releases)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

---

## 📌 Vision & Problem Statement

เมื่อทีมพัฒนามีขนาดใหญ่ขึ้น ปัญหาเรื่องมาตรฐานโค้ดและการละเมิดสถาปัตยกรรมมักจะเกิดขึ้นซ้ำๆ เช่น:
- 🔐 **Hardcoded Secrets:** มีคนเผลอวาง API Keys, AWS Credentials, หรือ Private Keys ลงในโค้ดดิบ
- 📁 **Structure & Mandatory Files:** ลืมสร้างไฟล์เอกสารสำคัญประจำองค์กร เช่น `README.md`, `.gitignore`, `Dockerfile`
- 🚨 **API Breaking Changes:** มีคนลบ Field หรือเปลี่ยน Type ใน API โดยไม่ตั้งใจ
- 📐 **Code Naming Standards:** ลืมทำตาม Naming Convention ของแต่ละ Framework

**ArchGuard** ทำหน้าที่เป็น **Engineering Policy Engine** ที่สแกนและบังคับใช้กฎทางวิศวกรรมแบบสากล ตั้งแต่บนเครื่องของ Developer ไปจนถึงระบบ CI/CD บน GitHub

---

## 🌟 Key Features

- 🚀 **CLI-First & Shift-Left:** สแกนโค้ดรวดเร็วปานสายฟ้าแลบบนเครื่องของ Developer ภายในเวลาไม่กี่มิลลิวินาที
- 🪄 **Interactive Setup Wizard (`archguard init`):** คำถามภาษาอังกฤษ 100% ปรับแต่ง `archguard.yaml` ให้ตรงตามสเปกโปรเจกต์อัตโนมัติ (รองรับ **Frontend**, **Backend**, **Full-Stack App**, และ **Library**)
- 🛡️ **Rich Built-in Policy Rules:**
  - `file-naming`: ตรวจมาตรฐานการตั้งชื่อไฟล์และโฟลเดอร์ตาม Regex (รองรับวงเล็บ `[id]` ของ Next.js)
  - `no-secrets`: สแกนหา AWS Access Keys, Private Keys, Bearer Tokens, และ Hardcoded Secrets
  - `required-files`: ตรวจบังคับความมีอยู่ของไฟล์สำคัญประจำองค์กร (เช่น `README.md`, `.gitignore`)
  - `openapi-exists`: ตรวจสอบความมีอยู่ของไฟล์เอกสารสัญญา API (OpenAPI/Swagger Spec)
- 📜 **Built-in Policy Catalog (`archguard rules`):** เรียกดูรายชื่อกฎ พารามิเตอร์ และ Default Severity บน Terminal ได้ทันที
- ⚓ **Git Pre-commit Hook Integration (`archguard install-hook`):** ติดตั้งระบบตรวจความถูกต้องอัตโนมัติก่อนสั่ง `git commit`
- 📊 **Multi-Format Reporting:** รองรับ Colored Console, JSON (`--format=json`), และ **SARIF v2.1.0 (`--format=sarif`)** สำหรับแสดงผลบน **GitHub Security Alerts** และไฮไลต์บรรทัดโค้ดใน **Pull Request**
- 📦 **Automated Cross-Platform Releases:** คอมไพล์ไฟล์ Binary สำเร็จรูปให้อัตโนมัติสำหรับ **Linux, macOS (Apple Silicon M1/M2/M3 & Intel), และ Windows**

---

## ⚡ Quick Start

### 1. Installation

#### Option A: Install via Go (Recommended for Developers)
```bash
go install github.com/jakkayy/archGuard/cmd/archguard@latest
```

#### Option B: Download Pre-compiled Binary
ดาวน์โหลดไฟล์สำเร็จรูปสำหรับ OS ของคุณจากหน้า [GitHub Releases](https://github.com/jakkayy/archGuard/releases/latest) แล้วนำไปวางไว้ใน System PATH

---

### 2. Initialize Project Configuration

เปิด Terminal ในโฟลเดอร์โปรเจกต์ของคุณแล้วรันคำสั่ง Interactive Wizard:
```bash
archguard init
```

---

### 3. Install Git Pre-Commit Hook (Optional but Recommended)

สั่งติดตั้งระบบตรวจสอบอัตโนมัติก่อนทำ `git commit` เพียงครั้งเดียว:
```bash
archguard install-hook
```

---

### 4. Run Policy Scan

```bash
archguard scan
```

---

## 💻 CLI Command Reference

| Command | Description | Flags |
| :--- | :--- | :--- |
| `archguard init` | เปิดหน้าต่าง Interactive Setup Wizard เพื่อสร้างไฟล์ `archguard.yaml` | `-f, --force` (เขียนทับไฟล์เดิม)<br>`-y, --non-interactive` (ข้ามคำถาม) |
| `archguard scan` | สแกนโปรเจกต์เพื่อตรวจสอบข้อผิดพลาดตามกฎใน `archguard.yaml` | `-c, --config <file>` (ไฟล์คอนฟิก)<br>`-f, --format <console\|json\|sarif>`<br>`--no-color` |
| `archguard rules` | แสดงรายการ Built-in Policy Rules ทั้งหมดพร้อมคำอธิบายและพารามิเตอร์ | N/A |
| `archguard install-hook` | ติดตั้งระบบ Pre-commit Hook อัตโนมัติไว้ที่ `.git/hooks/pre-commit` | N/A |

---

## ⚙️ Configuration Reference (`archguard.yaml`)

```yaml
version: "v1"

# ----------------------------------------------------
# 📂 Ignore Directories (ข้ามโฟลเดอร์ที่ไม่ต้องการตรวจ)
# ----------------------------------------------------
ignore:
  - "custom_build"
  - "tmp"

# ----------------------------------------------------
# 🛡️ Engineering Policies Rules
# ----------------------------------------------------
rules:
  # 1. File Naming Policy
  file-naming:
    enabled: true
    severity: WARNING
    pattern: '^[a-zA-Z0-9._\-\[\]\(\)]+$'

  # 2. Mandatory Files Check
  required-files:
    enabled: true
    severity: ERROR
    files:
      - "README.md"
      - ".gitignore"

  # 3. No Hardcoded Secrets Check
  no-secrets:
    enabled: true
    severity: ERROR

  # 4. OpenAPI Specification Check
  openapi-exists:
    enabled: false
    severity: ERROR
    path: "docs/openapi.json"
```

---

## 🤖 CI/CD Integration (GitHub Actions)

สร้างไฟล์ `.github/workflows/policy-check.yml` ในโปรเจกต์งานของคุณ:

```yaml
name: ArchGuard Engineering Policy Check

on: [push, pull_request]

jobs:
  archguard-scan:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Install ArchGuard Engine
        run: go install github.com/jakkayy/archGuard/cmd/archguard@latest

      - name: Run ArchGuard Policy Scan
        run: archguard scan

      - name: Generate & Upload SARIF Security Report
        run: |
          archguard scan --format=sarif > archguard-results.sarif
        continue-on-error: true

      - name: Upload SARIF to GitHub Security Tab
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: archguard-results.sarif
```

---

## 🏗️ Project Architecture Layout

```
archGuard/
├── .github/
│   └── workflows/            # GitHub Actions CI/CD & Automated Release Pipelines
├── cmd/
│   └── archguard/            # CLI Commands (main, root, init, scan, rules, install-hook)
├── internal/                 # Internal Private Core
│   ├── config/               # YAML Config Loader (archguard.yaml)
│   └── core/                 # Engine Orchestrator, Context & Issue Types
├── pkg/                      # Public SDK & Built-in Rules
│   ├── rule/                 # Rules (file-naming, no-secrets, openapi-exists, required-files)
│   └── reporter/             # Reporters (Console, JSON, SARIF)
├── docs/                     # Specifications & Documentation
└── archguard.yaml            # Default sample configuration file
```

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
