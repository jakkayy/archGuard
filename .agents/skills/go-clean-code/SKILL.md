---
name: go-clean-code
description: Enforces Clean Code guidelines, idiomatic Go naming conventions (Effective Go), SOLID architectural principles, package layout, and error handling standards for Go projects.
---

# 🐹 Go Clean Code & Standard Guidelines (Effective Go & Uber Go Style)

สเกิลนี้มีไว้เพื่อกำหนดและบังคับใช้มาตรฐานการเขียนโค้ดภาษา **Go** ให้ถูกต้องตามหลัก **Effective Go**, **Go Code Review Comments**, **Uber Go Style Guide**, และ **SOLID Principles** สำหรับโปรเจกต์ `ArchGuard`

---

## 📌 1. หลักการตั้งชื่อ (Idiomatic Go Naming Conventions)

### 1.1 Package Naming & Layout
- **ใช้คำเดียว ตัวพิมพ์เล็กทั้งหมด (Single-word, lowercase):** ห้ามใช้ `snake_case`, `camelCase`, หรือคำพหูพจน์ (Plural)
  - ✅ `package config`, `package core`, `package rule`, `package reporter`
  - ❌ `package Configs`, `package rule_engine`, `package common_utils`
- **`internal/` vs `pkg/` Scope:**
  - ใช้ `internal/` สำหรับ code ภายในที่ไม่อยากให้ external Go module อื่นนำไป import (Go Compiler จะบล็อกให้อัตโนมัติ)
  - ใช้ `pkg/` หรือ `sdk/` เฉพาะ Public Interfaces/SDK ที่ตั้งใจเปิดให้ภายนอกใช้งาน
- **หลีกเลี่ยง Package อเนกประสงค์ (Utility Sinks):** ห้ามสร้าง package ชื่อ `util`, `helper`, `common`, `base` เพราะเป็นการละเมิด Single Responsibility Principle

### 1.2 Exported vs Unexported Identifiers
- **Exported (Public):** ขึ้นต้นด้วยตัวพิมพ์ใหญ่ (e.g., `Rule`, `Engine`, `ScanResult`)
- **Unexported (Private):** ขึ้นต้นด้วยตัวพิมพ์เล็ก (e.g., `parseConfig`, `defaultRuleSet`)
- **ไม่มี `I` Prefix หน้า Interface:** ห้ามตั้งชื่อ Interface ขึ้นต้นด้วย `I` (เป็น anti-pattern ของ Java/C#)
  - ✅ `type Rule interface`, `type Reporter interface`
  - ❌ `type IRule interface`, `type IReporter interface`
- **Interface ที่มี Method เดียว:** ลงท้ายด้วย `-er` 
  - ✅ `type Runner interface { Run(...) }`
  - ✅ `type Reporter interface { Report(...) }`

### 1.3 Acronyms & Initialisms (คำย่อ)
- ต้องใช้ Case แบบเดียวกันตลอดทั้งคำย่อ (Consistent Casing)
  - ✅ `CLI`, `JSON`, `API`, `URL`, `ID`, `YAML`, `SARIF`
  - ✅ `JSONReporter`, `RuleID`, `ParseYAML`, `APIEndpoint`
  - ❌ `JsonReporter`, `RuleId`, `ParseYaml`, `ApiEndpoint`

### 1.4 Variable & Receiver Naming
- **Receiver Naming:** Receiver Name ใน Struct Method ต้องสั้น (1-2 อักขระ) สอดคล้องกับชื่อ Struct และ **ห้ามใช้ `this`, `self`, `me`** เด็ดขาด
  - ✅ `func (e *Engine) Run(...)`
  - ✅ `func (r *FileNamingRule) ID() string`
  - ❌ `func (self *Engine) Run(...)`
- **ขอบเขตตัวแปร (Variable Scope):** 
  - ตัวแปรใน Scope เล็ก (Loop / Function) ให้ใช้ชื่อสั้น (`ctx`, `err`, `cfg`, `w`, `r`)
  - ตัวแปรใน Scope ใหญ่ (Package level / Struct field) ให้ใช้ชื่อยาวสื่อความหมาย

### 1.5 Avoid Stuttering (ห้ามตั้งชื่อซ้ำกับชื่อ Package)
- เมื่อนำไปใช้ภายนอก ชื่อ package จะถูกพิมพ์นำหน้าเสมอ จึงห้ามตั้งชื่อ type/func ซ้ำซ้อน
  - ✅ `config.Load()` / `config.Config`
  - ❌ `config.LoadConfig()` / `config.ConfigLoader`
  - ✅ `rule.Rule` / `rule.Registry`
  - ❌ `rule.RuleInterface` / `rule.RuleEngine`

### 1.6 Constructor Functions
- หาก package มี Main Type เพียงตัวเดียว ให้ใช้ชื่อ constructor ว่า `New(...)`
  - ✅ `engine.New(cfg)` ➔ คืนค่า `*Engine`
- หาก package มีหลาย Types ให้ใช้ `New<Type>(...)`
  - ✅ `rule.NewFileNamingRule(...)`

---

## 🏗️ 2. สถาปัตยกรรม & Clean Code Principles (SOLID)

### 2.1 Package Organization (Standard Go Layout)
- `cmd/`: มีหน้าที่เพียง parse CLI arguments/flags และประสานงานเรียกใช้ package ภายใน ห้ามใส่ Business Logic ไว้ใน `cmd/`
- `internal/`: เก็บ Core Domain Logic และ Scanner Engine ภายในโปรเจกต์
- แต่ละ Package ต้องมีขอบเขตความรับผิดชอบเดียว (**Single Responsibility Principle**)

### 2.2 Interface & Dependency Inversion
- **Accept Interfaces, Return Structs:** ฟังก์ชันหรือ Constructor ควรรับพารามิเตอร์เป็น Interface และส่งคืน Concrete Struct (`*Struct`)
- **Interface Definition:** นิยาม Interface ฝั่งผู้เรียกใช้ (Consumer side) ไม่ใช่นิยามไว้ที่ Producer เสมอไป

### 2.3 Explicit Error Handling & Sentinel Errors
- ** sentinel Error Naming:** ขึ้นต้นด้วย `Err` (e.g. `var ErrConfigNotFound = errors.New(...)`)
- ** Error String Format:** ข้อความใน `errors.New()` ต้องเป็น **ตัวพิมพ์เล็กทั้งหมดและห้ามจบด้วยเครื่องหมายจุด (.)** ตามมาตรฐาน Go
  - ✅ `errors.New("config file not found")`
  - ❌ `errors.New("Config file not found.")`
- **Wrap Error พร้อม Context:** ใช้ `fmt.Errorf("context message: %w", err)`
  - ✅ `return fmt.Errorf("failed to load yaml config from %s: %w", path, err)`
  - ❌ `return err` (โดยไม่บอก context)
- **ห้าม Swallow Error:** ต้องตรวจสอบ `if err != nil` ทุกครั้ง
- **ห้ามใช้ `panic()` ใน Normal Control Flow:** ใช้ `panic()` เฉพาะ Unrecoverable Initialization Failure เท่านั้น

### 2.4 Context Propagation
- ฟังก์ชันหรือ Method ที่ทำงานเกี่ยวกับ I/O, File System scanning, หรือ long-running operations **ต้องรับ `context.Context` เป็นพารามิเตอร์ตัวแรกเสมอ**
  - ✅ `func (e *Engine) Run(ctx context.Context) (*ScanResult, error)`

---

## 🧪 3. การทดสอบและความสะอาดของโค้ด (Testing & Quality)

1. **Table-Driven Tests:** ใช้โครงสร้าง Table-Driven Test ของ Go ในการเขียน Unit Test เสมอ
2. **Zero Global State:** หลีกเลี่ยงการใช้ตัวแปรระดับ Global/Package (Global Mutables) ทุกกรณี
3. **Format & Lint Compliance:** โค้ดทั้งหมดต้องผ่าน `gofmt` และ `go vet` เสมอ
