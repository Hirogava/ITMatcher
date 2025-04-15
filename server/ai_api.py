from fastapi import FastAPI
import spacy
import uvicorn
from typing import Dict, List
from pydantic import BaseModel

app = FastAPI()

# Глобальные настройки
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
    "unix": "Unix"
}

STOP_WORDS = {
    "HARDSKILL": {
        # Убрали многие технические термины из стоп-слов
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

def normalize_skill(text: str, entity_type: str) -> str:
    """Нормализация названий навыков"""
    text = text.strip()
    lower_text = text.lower()
    
    # Применяем замены из словаря нормализации
    for pattern, replacement in SKILL_NORMALIZATION.items():
        if pattern in lower_text:
            return replacement
    
    # Специальные правила для технических навыков
    if entity_type == "HARDSKILL":
        if len(text) <= 3:
            return text.upper()
        elif text.isupper() and len(text) > 3:
            return text.capitalize()
    
    # Общие правила
    if len(text.split()) > 1:
        return ' '.join(word.capitalize() for word in text.split())
    return text.capitalize()

def extract_entities(doc, entity_type: str) -> List[str]:
    """Извлечение сущностей из документа с улучшенной фильтрацией"""
    valid_entities = set()
    skip_prefixes = {'http', 'www', 'github.com', 'habr', 'medium'}
    
    for ent in doc.ents:
        if ent.label_ == entity_type:
            original_text = ent.text.strip()
            lower_text = original_text.lower()
            
            # Пропускаем ссылки и нежелательные паттерны
            if any(lower_text.startswith(prefix) for prefix in skip_prefixes):
                continue
                
            # Пропускаем стоп-слова
            if lower_text in STOP_WORDS.get(entity_type, set()):
                continue
                
            # Нормализуем текст
            normalized = normalize_skill(original_text, entity_type)
            
            # Основные фильтры
            if (len(normalized) >= 2 
                and not normalized.isdigit()
                and not any(c in normalized for c in ['\n', ':', '/', '(', ')'])
                and any(c.isalpha() for c in normalized)):
                valid_entities.add(normalized)
    
    return postprocess_skills(list(valid_entities))

def postprocess_skills(skills: List[str]) -> List[str]:
    """Постобработка извлеченных навыков"""
    # Удаляем дубликаты и подстроки
    skills = sorted(list(set(skills)), key=lambda x: (-len(x), x))
    return [s for i, s in enumerate(skills) 
            if not any(s.lower() in longer.lower() for longer in skills[:i])]

class ResumeRequest(BaseModel):
    text: str

@app.post("/analyze")
def analyze(req: ResumeRequest) -> Dict[str, List[str]]:
    try:
        # Загружаем модель (должна быть предварительно обучена)
        nlp_ru = spacy.load("./output/model-best")
        
        # Предобработка текста
        text = ' '.join(req.text.split())  # Удаляем лишние пробелы
        
        # Анализ текста
        doc = nlp_ru(text)
        
        # Извлекаем сущности
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