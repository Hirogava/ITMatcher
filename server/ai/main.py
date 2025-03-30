import os
from openai import OpenAI

client = OpenAI(
    base_url="https://api.aimlapi.com/v1",

    # Insert your AIML API Key in the quotation marks instead of <YOUR_AIMLAPI_KEY>:
    api_key="cc816599d10744daa9316879a9b6f88a",  
)

with open('resume.txt', 'r', encoding="UTF-8") as f:
    resume = f.read()

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[
        {
            "role": "user",
            "content": f"Вытащи навыки из резюме, с разделением на soft скиллы и hard, указывая их в том же порядке, что они были описаны в резюме, выделяя только названия навыков, например: SQL, Python и тд: {resume}"
        },
    ],
)

message = response.choices[0].message.content

print(f"Assistant: {message}")