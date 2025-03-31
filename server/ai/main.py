import sys
from openai import OpenAI

client = OpenAI(
    base_url="https://openrouter.ai/api/v1",
    api_key="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",
)

try:
    resume = sys.stdin.read()

    completion = client.chat.completions.create(
    extra_body={},
    model="qwen/qwen2.5-vl-32b-instruct:free",
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

    print(completion.choices[0].message.content.encode('utf-8').decode('utf-8'))
except Exception as e:
    print(f"Error: {e}", file=sys.stderr)
    sys.exit(1)