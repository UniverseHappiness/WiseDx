<p align="center">
  <picture>
    <img src="./frontend/src/assets/img/wisedx.png" alt="WiseDx Logo" height="120"/>
  </picture>
</p>

<p align="center">
    <a href="https://github.com/Tencent/WeKnora/blob/main/LICENSE">
        <img src="https://img.shields.io/badge/License-MIT-ffffff?labelColor=d4eaf7&color=2e6cc4" alt="License">
    </a>
    <a href="./CHANGELOG.md">
        <img alt="Version" src="https://img.shields.io/badge/version-0.2.10-2e6cc4?labelColor=d4eaf7">
    </a>
</p>

<p align="center">
| <b>English</b> | <a href="./README_CN.md"><b>简体中文</b></a> | <a href="./README_JA.md"><b>日本語</b></a> |
</p>

<p align="center">
  <h4 align="center">

  [Overview](#-overview) • [Key Features](#-key-features) • [Medical Capabilities](#-medical-capabilities) • [Getting Started](#-getting-started) • [Acknowledgements](#-acknowledgements)

  </h4>
</p>

# 💡 WiseDx - Professional Medical RAG Q&A and Retrieval Framework

## 📌 Overview

**WiseDx** is an LLM-powered framework specifically engineered for medical knowledge understanding and Retrieval-Augmented Generation (RAG). It is designed to tackle the unique challenges of the medical domain, such as complex document structures (clinical guidelines, medical records, research papers), highly specialized terminology, and the critical need for evidence-based accuracy.

By integrating multimodal medical document parsing, medical-domain semantic indexing, Evidence-Based Medicine (EBM) knowledge graph enhancement, and Agentic reasoning, WiseDx provides professional, precise, and traceable medical knowledge services for healthcare professionals and researchers.

## ✨ Key Features

- **🤖 Medical Agent Mode**: Based on the ReACT architecture, it supports autonomous planning and reasoning. It can call tools like medical guideline retrieval, academic web search, and specialized calculators to deliver in-depth medical summary reports.
- **🔍 Multimodal Medical Parsing**: Deep support for PDF, Word, and medical imaging reports, with precise extraction of medical tables, charts, and professional terminology.
- **🧠 Evidence-Based Enhancement**: Incorporates medical knowledge graphs into the RAG workflow to ensure retrieval results align with clinical logic, enhancing the professional authority of the responses.
- **⚡ Hybrid Retrieval Engine**: Combines keyword matching (with MeSH support), clinical semantic vector retrieval, and relationship-based recall to significantly improve recall accuracy.
- **🔒 Secure Private Deployment**: Supports local and private cloud deployment, ensuring absolute security for patient privacy and sensitive medical data.

## 📊 Medical Application Scenarios

| Scenario | Application | Core Value |
|---------|-------------|------------|
| **Clinical Decision Support** | Guideline querying, complex case analysis, differential diagnosis assistance | Helps clinicians quickly access evidence, reducing misdiagnosis risks. |
| **Medical Research Analysis** | Semantic search across literature, review generation, tracking front-line research | Accelerates literature review and helps discover potential research links. |
| **Pharmacy Knowledge Service** | Drug interaction check, contraindication verification, drug manual retrieval | Enhances clinical medication safety with real-time consultation. |
| **Medical Compliance Review** | Quality indicator checking, medical record standardization, regulatory search | Automates quality management and reduces compliance risks. |

## 🧩 Medical Specialized Capabilities

- **✅ Terminology Mapping**: Supports mapping and retrieval across standard sets like ICD-10, MeSH, and SNOMED CT.
- **✅ Evidence Traceability**: Every generated answer can be traced back to the original medical literature, guideline, or patient record segment.
- **✅ Multimodal Understanding**: Deeply parses documents containing ECGs, medical image descriptions, and laboratory test images.
- **✅ Clinical Reasoning**: The Agent possesses clinical thinking capabilities, performing chain-of-thought reasoning based on observed symptoms and signs.

## 🚀 Getting Started

### 🛠 Prerequisites

Ensure the following tools are installed:
* [Docker](https://www.docker.com/)
* [Docker Compose](https://docs.docker.com/compose/)
* [Git](https://git-scm.com/)

### 📦 Installation

```bash
# Clone the repository
git clone https://github.com/Tencent/WeKnora.git
cd WeKnora

# Configure environment variables
cp .env.example .env

# Start services
./scripts/start_all.sh
```

For more details, please refer to the [Developer Guide](./docs/开发指南.md).

## 🙏 Acknowledgements

The core architecture and technical implementation of **WiseDx** are deeply inspired by and built upon the [**WeKnora**](https://github.com/Tencent/WeKnora) open-source framework.

WeKnora is an exceptional RAG framework developed by the Tencent WeChat Dialog Open Platform team. Its modular design, powerful Agent capabilities, and support for various vector databases provided the solid technical foundation for WiseDx's vertical application in the medical field. We sincerely thank the WeKnora team and all contributors for their open-source spirit.

## 📄 License

This project is licensed under the [MIT License](./LICENSE).
