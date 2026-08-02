# 🛡️ ArchGuard

> **An Open-Source Engineering Policy Engine (ESLint for Engineering Policies)**
> 
> *เปลี่ยนกฎการเขียนโค้ดและสถาปัตยกรรมขององค์กร ให้กลายเป็นการตรวจสอบอัตโนมัติ (Automated Engineering Policy Engine)*

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

---

## 📌 Vision & Problem Statement

เมื่อทีมพัฒนามีขนาดใหญ่ขึ้น ปัญหาเรื่องมาตรฐานโค้ดและการละเมิดสถาปัตยกรรมมักจะเกิดขึ้นซ้ำๆ เช่น:
- **API Breaking Changes:** มีคนลบ Field หรือเปลี่ยน Type ใน API โดยไม่ตั้งใจ ส่งผลให้ระบบอื่นพัง
- **Architecture Violation:** เขียนโค้ดข้าม Layer (เช่น Controller เรียก Database โดยตรง)
- **Code Governance:** ลืมทำตาม Naming Convention หรือลืมสร้างไฟล์ Specification สำคัญ

**ArchGuard** ทำหน้าที่เป็น **Policy Engine** ที่สแกนและบังคับใช้กฎทางวิศวกรรม (Engineering Policies) ตั้งแต่บนเครื่องของ Developer (`archguard scan`) ไปจนถึงขั้นตอน CI/CD ก่อนจะ Merge โค้ดเข้าสู่ระบบจริง

---

## 🌟 Key Features

- 🚀 **CLI-First & Shift-Left:** รันสแกนบนเครื่อง Developer ได้ภายในไม่กี่มิลลิวินาที
- 🔌 **Policy-First & Plugin Architecture:** Core Engine ปราศจาก Hardcoded Business Rules และรองรับ Plugin ในอนาคต
- 🎨 **Rich Terminal Output:** แสดงผลการตรวจด้วยสีสัน แยกสัญลักษณ์ `🚨 ERROR`, `⚠️ WARN`, `ℹ️ INFO` ชัดเจน
- 📄 **Multiple Output Formats:** รองรับการแสดงผลทั้ง CLI Console และ **JSON** สำหรับนำไปใช้ใน CI/CD Pipeline
- 🛡️ **CI/CD Native:** คืนค่า Exit Code 1 อัตโนมัติเมื่อพบปัญหาระดับ `ERROR` สั่งบล็อก Pull Request ที่ไม่ผ่านมาตรฐานได้ทันที

---

## ⚡ Quick Start

### 1. Build Executable Binary
```bash
go build -o bin/archguard ./cmd/archguard
```

### 2. Initialize Configuration (`archguard.yaml`)
สร้างไฟล์ตั้งค่ากฎคอนฟิกตั้งต้นในไดเรกทอรีโปรเจกต์:
```bash
./bin/archguard init
```

### 3. Run Policy Scan
รันสแกนโปรเจกต์:
```bash
./bin/archguard scan
```

---

## ⚙️ Configuration Reference (`archguard.yaml`)

ตัวอย่างการตั้งค่าในไฟล์ `archguard.yaml`:

```yaml
version: "v1"
rules:
  # Rule 1: File Naming Convention Check
  file-naming:
    enabled: true
    severity: WARNING         # Options: ERROR, WARNING, INFO
    pattern: "^[a-z0-9._-]+$" # Regex rule for filename formatting

  # Rule 2: Required OpenAPI Spec File Check
  openapi-exists:
    enabled: true
    severity: ERROR           # Fails scan if spec file is missing
    path: "docs/archguard_bot_spec.md"
```

---

## 💻 CLI Commands & Options

### `archguard init`
สร้างไฟล์ `archguard.yaml` ตั้งต้น
- `-f, --force`: บังคับเขียนทับไฟล์ `archguard.yaml` เดิมที่มีอยู่

### `archguard scan`
สแกนโปรเจกต์เพื่อตรวจสอบข้อผิดพลาด
- `-c, --config <file>`: ระบุไฟล์คอนฟิก (Default: `archguard.yaml`)
- `-f, --format <type>`: ระบุรูปแบบรายงาน (`console` หรือ `json`) (Default: `console`)
- `--no-color`: ปิดการแสดงสีบน Terminal

#### Example: Output as JSON
```bash
./bin/archguard scan --format=json
```

---

## 🤖 CI/CD Integration (GitHub Actions)

สร้างไฟล์ `.github/workflows/policy-check.yml`:

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
          go-version: '1.26'

      - name: Build & Run ArchGuard Scan
        run: |
          go build -o bin/archguard ./cmd/archguard
          ./bin/archguard scan
```

---

## 🏗️ Project Architecture Layout

```
archguard/
├── cmd/
│   └── archguard/            # CLI Commands (main, root, init, scan)
├── internal/                 # Internal Private Core
│   ├── config/               # YAML Config Loader
│   └── core/                 # Engine Orchestrator, Context & Result Types
├── pkg/                      # Public SDK & Built-in Rules
│   ├── rule/                 # Rule Interface & Built-in Implementations
│   └── reporter/             # Console & JSON Formatter Reporters
├── docs/                     # Specifications & Documentation
└── archguard.yaml            # Sample configuration file
```

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
