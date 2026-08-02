# ArchGuard Project Rules

## Git Commit Guidelines
- หลังจากทำแต่ละ Task ใน Implementation Plan เสร็จเรียบร้อย ให้ทำการ Stage โค้ดและ Git Commit ทันที
- ใช้รูปแบบ **Conventional Commits** ที่กระชับเพียง **1 บรรทัด** (ไม่เกิน 72 ตัวอักษร):
  - `feat(<scope>): <คำอธิบายสั้นๆ>` (สำหรับฟีเจอร์/ความสามารถใหม่)
  - `fix(<scope>): <คำอธิบายสั้นๆ>` (สำหรับแก้ไขบั๊ก)
  - `docs(<scope>): <คำอธิบายสั้นๆ>` (สำหรับเอกสาร)
  - `chore(<scope>): <คำอธิบายสั้นๆ>` (สำหรับงานตั้งค่า/dependencies)
  - `test(<scope>): <คำอธิบายสั้นๆ>` (สำหรับการทดสอบ)
- ตัวอย่าง: `feat(core): define scan context and issue result types`
