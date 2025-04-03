import sys
import json
from openai import OpenAI

client = OpenAI(
    base_url="https://openrouter.ai/api/v1",
    api_key="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",
)

def sanitize_text(text):
    """Remove invalid Unicode surrogates."""
    return text.encode('utf-16', 'surrogatepass').decode('utf-16', 'ignore')

try:
    resume = sys.stdin.read()
    resume = sanitize_text(resume)

    completion = client.chat.completions.create(
        extra_body={},
        model="deepseek/deepseek-chat-v3-0324:free",
        messages=[
            {
                "role": "user",
                "content": (
                        "Проанализируй резюме и выдели из него навыки, разделяя их на две категории: Hard Skills и Soft Skills. "
                        "Выведи результат в следующем формате:\n\n"
                        "**Hard Skills:**\n"
                        "1. [Навык 1]\n"
                        "2. [Навык 2]\n"
                        "...\n\n"
                        "**Soft Skills:**\n"
                        "1. [Навык 1]\n"
                        "2. [Навык 2]\n"
                        "...\n\n"
                        "Важно:\n"
                        "- Указывай только конкретные названия навыков.\n"
                        "- Не группируй навыки в категории или подтемы (например, вместо \"Программирование (Python, SQL)\" выдели \"Python\" и \"SQL\" как отдельные навыки).\n"
                        "- Сохраняй порядок, в котором навыки упоминаются в резюме.\n"
                        "- Не добавляй комментарии или пояснения.\n\n"
                        "- Не добавляй ничего лишнего. Выведи только результат в указанном формате.\n"
                        "- Не добавляй примечания или пояснения.\n"
                        "Резюме:\n"
                        f"{resume}"
                )
            }
        ]
    )

    skills_text = completion.choices[0].message.content.strip()
    hard_skills = []
    soft_skills = []
    
    current_category = None
    for line in skills_text.split('\n'):
        line = line.strip()
        if line.startswith("**Hard Skills:**"):
            current_category = hard_skills
        elif line.startswith("**Soft Skills:**"):
            current_category = soft_skills
        elif line and line[0].isdigit() and current_category is not None:
            skill = line.split(". ", 1)[-1]
            current_category.append(skill)
    
    result = {
        "hard_skills": hard_skills,
        "soft_skills": soft_skills
    }
    
    print(json.dumps(result, indent=4, ensure_ascii=False))

except Exception as e:
    print(f"Error: {e}", file=sys.stderr)
    sys.exit(1)