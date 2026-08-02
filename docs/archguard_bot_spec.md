# 🛡️ ArchGuard Specification

> **ระบบ Open-source สำหรับกำหนดและบังคับใช้มาตรฐานการพัฒนาซอฟต์แวร์ (Engineering Policy Engine)**
> 
> **Tagline:** เปลี่ยนกฎการเขียนโค้ดขององค์กร ให้กลายเป็นกฎที่ตรวจสอบได้อัตโนมัติ (ESLint for Engineering Policies)

---

## 📌 1. วิสัยทัศน์ (Vision) & ปัญหาที่ต้องการแก้ (Problem Statement)

### 1.1 วิสัยทัศน์
ArchGuard ไม่ใช่เพียงแค่ GitHub Bot ที่คอยแสดงความคิดเห็นใน Pull Request แต่เป็น **Policy Engine** ที่ช่วยให้องค์กรสามารถนิยาม "กฎทางวิศวกรรม" (Engineering Policies) และบังคับใช้กฎเหล่านั้นได้อัตโนมัติ ตั้งแต่ในระดับเครื่องของ Developer (Local CLI / Pre-commit) ไปจนถึงขั้นตอน CI/CD ก่อนจะ Merge โค้ดเข้าสู่ระบบ

### 1.2 ปัญหาในทีมพัฒนาจริง (Pain Points)
1. **Reviewer เสียเวลากับเรื่องเดิมๆ:** ต้องคอยจับผิดเรื่องเดิมทุก PR เช่น "ลืมเขียน Unit Test", "ทำไม Controller ต่อ Database โดยตรง"
2. **API Breaking Changes พังระบบอื่น:** มีการลบ Field หรือเปลี่ยน Type ใน API โดยไม่ตั้งใจ ทำให้ Frontend หรือ Mobile App เวอร์ชันเก่าค้างและพัง
3. **Architecture ถูกละเมิด:** เมื่อทีมสเกลใหญ่ขึ้น กฎ Architecture Layer Boundary ถูกละเมิดจนระบบเละ
4. **กฎถูกลืมใน Wiki:** Documentation กฎของทีมมีอยู่จริง แต่ไม่มีใครเปิดอ่านเพราะไม่มีระบบตรวจสอบอัตโนมัติ

---

## 🏗️ 2. แนวคิดการออกแบบสถาปัตยกรรม (Design Principles)

1. **Policy First:** ทุกอย่างคือ Policy (Architecture, API, Security, Testing, Infrastructure, Naming) Core จะไม่ผูกติดกับ Logic ของกฎใดกฎหนึ่ง
2. **Plugin Architecture:** Core มีหน้าที่เพียงอ่าน Config ➔ โหลด Plugin ➔ ส่ง Context ➔ รวมผลลัพธ์ ➔ ส่งออก Report
3. **Multi-Language & Multi-Platform:** Core เขียนด้วย **Go** (Single Binary, Fast Execution) ออกแบบให้สแกนและรองรับภาษา Go, TypeScript, Java, Python, Rust ฯลฯ ผ่าน Plugin และรองรับ GitHub, GitLab, Bitbucket ผ่าน Adapters
4. **Shift-Left Developer Experience (DX):** นักพัฒนารัน `archguard scan` หรือใช้ `pre-commit` บนเครื่องตัวเองได้ทันทีในเวลาไม่กี่วินาที
5. **Standardized Outputs:** รองรับการแสดงผลทั้ง CLI Console (Colored), JSON, HTML และ **SARIF** (Static Analysis Results Interchange Format) สำหรับเชื่อมต่อ GitHub Security / CI Tools

---

## 💻 3. สถาปัตยกรรมระบบ (System Architecture)

```mermaid
graph TD
    Dev[ Developer / CLI ] -->|archguard scan| Core[ ArchGuard Core Engine (Go) ]
    CI[ GitHub Action / GitLab CI ] -->|archguard scan| Core
    
    subgraph Core Engine Execution Flow
        Core --> ConfigLoader[ Config Loader (archguard.yaml) ]
        Core --> PluginLoader[ Plugin & Rule Engine ]
        
        subgraph Rules & Plugins
            PluginLoader --> Rule1[ OpenAPI Spec Rule ]
            PluginLoader --> Rule2[ Architecture Layer Rule ]
            PluginLoader --> Rule3[ Naming Convention Rule ]
            PluginLoader --> Rule4[ Test Requirement Rule ]
        end
        
        Core --> ReportEngine[ Report Engine ]
    end

    ReportEngine --> OutputConsole[ CLI Console Output ]
    ReportEngine --> OutputSARIF[ SARIF / JSON File ]
    ReportEngine --> OutputPR[ GitHub Check / PR Comment ]
```

---

## 🛠️ 4. เทคโนโลยีที่ใช้ในการพัฒนา (Tech Stack)

* **Core Engine:** **Go 1.26+** (Compile เป็น Single Binary ทำงานเร็ว ปราศจาก Runtime Dependency)
* **CLI Framework:** `github.com/spf13/cobra`
* **Configuration:** `gopkg.in/yaml.v3`
* **Terminal Formatting:** `github.com/fatih/color`
* **Plugin Mechanism (Future Roadmap):** WebAssembly (WASM via Extism) / gRPC (HashiCorp `go-plugin`)
* **Multi-Language Parsing:** AST Parsing / Tree-sitter integration

---

## 📁 5. โครงสร้างโปรเจกต์ (Project Layout)

```
archguard/
├── cmd/
│   └── archguard/            # CLI Main Entrypoint
│       └── main.go
├── pkg/
│   ├── config/               # Configuration Loader (archguard.yaml)
│   ├── core/                 # Scanner Engine, Context & Issue Types
│   ├── rule/                 # Rule Interface & Built-in Rule Implementations
│   └── reporter/             # Console, JSON, SARIF Formatter
├── docs/                     # Specifications & Architecture Docs
├── examples/                 # Sample Configuration & Project Templates
├── archguard.yaml            # Sample configuration file
├── go.mod
└── go.sum
```

---

## 🗓️ 6. แผนการพัฒนา (Roadmap)

### 🔹 v0.1 (MVP - CLI & Core Foundation)
- [ ] สถาปัตยกรรม Go Project & Command Structure (`archguard scan`, `archguard init`)
- [ ] Config Loader อ่านไฟล์ `archguard.yaml`
- [ ] Rule Interface & Core Engine Runner
- [ ] Built-in Rules: `file-naming` และ `openapi-exists`
- [ ] Console Formatter แสดงผลการตรวจบน Terminal

### 🔹 v0.2 (Rule Expansion & Layer Boundary)
- [ ] OpenAPI Breaking Change Detection
- [ ] Go / TypeScript AST Layer Boundary Rule
- [ ] JSON & SARIF Output Reporters

### 🔹 v0.3 (Git Platform & CI Integration)
- [ ] GitHub Action Component
- [ ] GitHub App / PR Commenter & Status Checks
- [ ] GitLab CI Support

### 🔹 v0.4 (Plugin SDK & Extensibility)
- [ ] WASM / gRPC Plugin SDK
- [ ] Rule Testing Framework

### 🔹 v1.0 (Public Release & Marketplace)
- [ ] Stable Plugin API
- [ ] Community Plugin Marketplace
- [ ] Homebrew / Docker / Cross-platform Binaries
- [ ] Documentation Website

---
*เอกสารปรับปรุงล่าสุด: 2026-08-02*
