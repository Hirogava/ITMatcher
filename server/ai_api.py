from fastapi import FastAPI
import spacy
import uvicorn
from typing import Dict, List
from pydantic import BaseModel

app = FastAPI()

SKILL_NORMALIZATION = {
    "golang": "Go",
    "postgresql": "PostgreSQL",
    "mysql": "MySQL",
    "mongodb": "MongoDB",
    "redis": "Redis",
    "grpc": "gRPC",
    "rest api": "REST API",
    "websockets": "WebSockets",
    "github actions": "GitHub Actions",
    "gitlab ci/cd": "GitLab CI/CD",
    "bash": "Bash",
    "linux": "Linux",
    "unix": "Unix",
    "power bi": "Power BI",
    "tableau": "Tableau",
    "c++": "C++",
    "go": "Go",
    "python": "Python",
    "java": "Java",
    "javascript": "JavaScript",
    "typescript": "TypeScript",
    "docker": "Docker",
    "kubernetes": "Kubernetes",
    "aws": "AWS",
    "azure": "Azure",
    "gcp": "GCP",
    "sql": "SQL",
    "nosql": "NoSQL",
    "node.js": "Node.js",
    "react": "React",
    "angular": "Angular",
    "vue": "Vue.js",
    "математическое мышление": "Математическое мышление",
    "data analysis": "Data Analysis",
    "machine learning": "Machine Learning",
    "r": "R",
    "scala": "Scala",
    "tensorflow": "TensorFlow",
    "pytorch": "PyTorch",
    "hadoop": "Hadoop",
    "spark": "Spark",
    "excel": "Excel",
    "agile": "Agile",
    "scrum": "Scrum",
    "jira": "JIRA",
    "confluence": "Confluence",
    "ci/cd": "CI/CD",
    "jenkins": "Jenkins",
    "ansible": "Ansible",
    "terraform": "Terraform",
    "1c": "1C",
    "low-code": "Low-code",
    "outsystems": "OutSystems",
    "mendix": "Mendix",
    "appian": "Appian",
    "powerapps": "PowerApps",
    "salesforce": "Salesforce",
    "sap": "SAP",
    "oracle": "Oracle",
    "graphql": "GraphQL",
    "ruby": "Ruby",
    "php": "PHP",
    "perl": "Perl",
    "swift": "Swift",
    "objective-c": "Objective-C",
    "kotlin": "Kotlin",
    "rust": "Rust",
    "dart": "Dart",
    "flutter": "Flutter",
    "bootstrap": "Bootstrap",
    "sass": "SASS",
    "less": "LESS",
    "webpack": "Webpack",
    "babel": "Babel",
    "express.js": "Express.js",
    "spring boot": "Spring Boot",
    "hibernate": "Hibernate",
    "laravel": "Laravel",
    "symfony": "Symfony",
    "flask": "Flask",
    "fastapi": "FastAPI",
    "asp.net": "ASP.NET",
    "pandas": "Pandas",
    "numpy": "NumPy",
    "scipy": "SciPy",
    "matplotlib": "Matplotlib",
    "seaborn": "Seaborn",
    "plotly": "Plotly",
    "d3.js": "D3.js",
    "data mining": "Data Mining",
    "predictive modeling": "Predictive Modeling",
    "optimization": "Optimization",
    "quantitative analysis": "Quantitative Analysis",
    "statistical modeling": "Statistical Modeling",
    "regression analysis": "Regression Analysis",
    "time series analysis": "Time Series Analysis",
    "c#": "C#",
    "c": "C",
    "matlab": "MATLAB",
    "simulink": "Simulink",
    "sas": "SAS",
    "spss": "SPSS",
    "stata": "Stata",
    "julia": "Julia",
    "opencv": "OpenCV",
    "keras": "Keras",
    "mxnet": "MXNet",
    "caffe": "Caffe",
    "blockchain": "Blockchain",
    "smart contracts": "Smart Contracts",
    "solidity": "Solidity",
    "hyperledger": "Hyperledger",
    "ethereum": "Ethereum",
    "quantum computing": "Quantum Computing",
    "qiskit": "Qiskit",
    "bioinformatics": "Bioinformatics",
    "genomics": "Genomics",
    "protobuf": "Protobuf",
    "swiftui": "SwiftUI",
    "minio": "MinIO",
    "koobiq": "Koobiq",
    "zig": "Zig",
    "nim": "Nim",
    "crystal": "Crystal",
    "ocaml": "OCaml",
    "elm": "Elm",
    "webassembly": "WebAssembly",
    "svelte": "Svelte",
    "next.js": "Next.js",
    "nuxt.js": "Nuxt.js",
    "nestjs": "NestJS",
    "neo4j": "Neo4j",
    "clickhouse": "ClickHouse",
    "airflow": "Airflow",
    "mlflow": "MLflow",
    "langchain": "LangChain",
    "llms": "LLMs",
    "prompt engineering": "Prompt Engineering",
    "mlops": "MLOps",
    "devsecops": "DevSecOps",
    "iac": "IaC",
    "pulumi": "Pulumi",
    "xgboost": "XGBoost",
    "lightgbm": "LightGBM",
    "catboost": "CatBoost",
    "huggingface transformers": "HuggingFace Transformers",
    "spacy": "spaCy",
    "nltk": "NLTK",
    "stanza": "Stanza",
    "ocr": "OCR",
    "yolo": "YOLO",
    "gans": "GANs",
    "vae": "VAE",
    "jax": "JAX",
    "kafka": "Kafka",
    "splunk": "Splunk",
    "datadog": "Datadog",
    "grafana": "Grafana",
    "opentelemetry": "OpenTelemetry",
    "metasploit": "Metasploit",
    "burp suite": "Burp Suite",
    "wireshark": "Wireshark",
    "gdpr": "GDPR",
    "hipaa": "HIPAA",
    "pci dss": "PCI DSS",
    "sap hana": "SAP HANA",
    "servicenow": "ServiceNow",
    "qlikview": "QlikView",
    "qlik sense": "Qlik Sense",
    "abap": "ABAP",
    "t-sql": "T-SQL",
    "pl/sql": "PL/SQL",
    "openshift": "OpenShift",
    "helm": "Helm",
    "istio": "Istio",
    "webrtc": "WebRTC",
    "ffmpeg": "FFmpeg",
    "three.js": "Three.js",
    "unity": "Unity",
    "unreal engine": "Unreal Engine",
    "blender": "Blender",
    "latex": "LaTeX",
    "gherkin": "Gherkin",
    "cucumber": "Cucumber",
    "testrail": "TestRail",
    "pytest": "PyTest",
    "databricks": "Databricks",
    "snowflake": "Snowflake",
    "redshift": "Redshift",
    "figma": "Figma",
    "sketch": "Sketch",
    "adobe xd": "Adobe XD",
    "wordpress": "WordPress",
    "joomla": "Joomla",
    "drupal": "Drupal",
    "bitrix": "Bitrix",
    "opencart": "OpenCart",
    "yii2": "Yii2",
    "laravel5": "Laravel 5",
    "nginx": "Nginx",
    "gimp": "GIMP",
    "вёрстка сайтов": "Вёрстка сайтов",
    "оптимизация кода": "Оптимизация кода",
    "архитектура кода": "Архитектура кода",
    "разработка фреймворка": "Разработка фреймворков",
    "написание скриптов": "Написание скриптов",
    "формирование баз данных": "Формирование баз данных",
    "создание cms": "Создание CMS",
    "создание crm": "Создание CRM",
    "установка на cms": "Установка на CMS"
}

SOFT_SKILL_NORMALIZATION = {
    "коммуникабельность": "Коммуникабельность",
    "лидерство": "Лидерство",
    "работа в команде": "Работа в команде",
    "критическое мышление": "Критическое мышление",
    "адаптивность": "Адаптивность",
    "решение проблем": "Решение проблем",
    "управление временем": "Тайм-менеджмент",
    "креативность": "Креативность",
    "эмоциональный интеллект": "Эмоциональный интеллект",
    "навыки презентации": "Навыки презентации",
    "переговоры": "Навыки переговоров",
    "организованность": "Организованность",
    "инициативность": "Инициативность",
    "стрессоустойчивость": "Стрессоустойчивость",
    "гибкость": "Гибкость",
    "мотивация": "Мотивация",
    "внимание к деталям": "Внимание к деталям",
    "навыки слушания": "Активное слушание",
    "управление конфликтами": "Управление конфликтами",
    "самомотивация": "Самомотивация",
    "навыки наставничества": "Наставничество",
    "управление проектами": "Управление проектами",
    "клиентоориентированность": "Клиентоориентированность",
    "управление изменениями": "Управление изменениями",
    "навыки делегирования": "Делегирование",
    "навыки коучинга": "Коучинг",
    "навыки фасилитации": "Фасилитация",
    "навыки убеждения": "Убеждение",
    "эмпатия": "Эмпатия",
    "культурная осведомленность": "Культурная осведомленность",
    "аналитические способности": "Аналитические способности",
    "приоритизация": "Приоритизация задач",
    "обучаемость": "Обучаемость",
    "ответственность": "Ответственность",
    "культурная адаптивность": "Культурная адаптивность",
    "чтение технической литературы": "Чтение технической литературы",
    "саморазвитие": "Саморазвитие",
    "достиженческое мышление": "Достиженческое мышление",
    "инициативность": "Инициативность",
    "проактивность": "Проактивность",
    "ассертивность": "Ассертивность",
    "публичные выступления": "Публичные выступления",
    "цифровой этикет": "Цифровой этикет"
}

STOP_WORDS = {
    "HARDSKILL": {
        "code", "работа", "review", "source", "best", "practices", "код",
        "github.com/username", "английский", "unit", "habr", "medium", "github",
        "abc", "hard", "intermediate", "open", "actions", "a/b", "adobe",
        "senior", "junior", "team", "project", "experience", "years", "лет", "год", "года",
        "company", "работал", "разработка", "development", "система", "system", "программа",
        "application", "инструмент", "tool", "platform", "платформа", "сервис", "service",
        "интеграция", "integration", "оптимизация", "optimization", "документация", "documentation",
        "тестирование", "testing", "deploy", "развертывание", "backend-разработка", "frontend",
        "api", "интерфейс", "interface", "база", "database", "обучение", "training", "курс",
        "course", "университет", "university", "магистратура", "бакалавриат", "роль", "role",
        "position", "должность", "задача", "task", "процесс", "process", "метод", "method",
        "подход", "approach", "результат", "result", "проект", "projects", "клиент", "client",
        "пользователь", "user", "команда", "team", "лидер", "leader", "менеджмент", "management",
        "инфраструктура", "infrastructure", "http", "url", "www", "com", "ru", "org", "рф"
    },
    "SOFTSKILL": {
        "работа", "производительности", "best", "practices", "rate", "retention",
        "опыт", "experience", "год", "лет", "years", "практика", "practice", "решение", "solution",
        "задачи", "tasks", "проекты", "projects", "команда", "team", "коллеги", "colleagues",
        "клиенты", "clients", "взаимодействие", "interaction", "коммуникация с", "communication with",
        "дедлайн", "deadline", "качество", "quality", "результат", "result", "цель", "goal",
        "план", "plan", "стратегия", "strategy", "поддержка", "support", "обучение", "training",
        "развитие", "development", "рост", "growth", "процесс", "process", "условия", "conditions",
        "ситуация", "situation", "контекст", "context", "знание", "knowledge", "умение", "ability",
        "навык", "skill", "достижение", "achievement", "успех", "success", "ошибка", "error",
        "проблема", "problem", "анализ", "analysis", "оценка", "evaluation"
    }
}

def preprocess_text(text: str) -> str:
    text = text.replace('(', ' (').replace(')', ') ')
    lines = [line.strip() for line in text.split('\n') if line.strip()]
    return '. '.join(lines) + '.'

def normalize_skill(text: str, entity_type: str) -> str:
    text = text.strip()
    lower_text = text.lower()
    
    if entity_type == "HARDSKILL" and lower_text in SKILL_NORMALIZATION:
        return SKILL_NORMALIZATION[lower_text]
    elif entity_type == "SOFTSKILL" and lower_text in SOFT_SKILL_NORMALIZATION:
        return SOFT_SKILL_NORMALIZATION[lower_text]
    
    if entity_type == "HARDSKILL":
        if len(text) <= 3:
            return text.upper()
        elif text.isupper() and len(text) > 3:
            return text.capitalize()
    
    if len(text.split()) > 1:
        return ' '.join(word.capitalize() for word in text.split())
    return text.capitalize()

def extract_entities(doc, entity_type: str) -> List[str]:
    valid_entities = set()
    skip_prefixes = {'http', 'www', 'github.com', 'habr', 'medium'}
    
    for ent in doc.ents:
        if ent.label_ == entity_type:
            original_text = ent.text.strip()
            lower_text = original_text.lower()
            
            if any(lower_text.startswith(prefix) for prefix in skip_prefixes):
                continue
                
            if lower_text in STOP_WORDS.get(entity_type, set()):
                continue
                
            normalized = normalize_skill(original_text, entity_type)
            
            if (len(normalized) >= 2 
                and not normalized.isdigit()
                and not any(c in normalized for c in ['\n', ':', '/', '(', ')'])
                and any(c.isalpha() for c in normalized)):
                valid_entities.add(normalized)
    
    return postprocess_skills(list(valid_entities))

def postprocess_skills(skills: List[str]) -> List[str]:
    skills = list(set(skills))
    skills.sort(key=lambda x: (-len(x), x.lower()))
    return [s for i, s in enumerate(skills) 
            if not any(s.lower() in longer.lower() for longer in skills[:i])]

class ResumeRequest(BaseModel):
    text: str

@app.post("/analyze")
def analyze(req: ResumeRequest) -> Dict[str, List[str]]:
    try:
        nlp_ru = spacy.load("./output/model-best")
        text = preprocess_text(req.text)
        doc = nlp_ru(text)
        
        hardskills = extract_entities(doc, "HARDSKILL")
        softskills = extract_entities(doc, "SOFTSKILL")
        
        return {
            "hard_skills": hardskills,
            "soft_skills": softskills
        }
        
    except Exception as e:
        return {"error": str(e)}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8001)