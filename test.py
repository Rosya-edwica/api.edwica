import requests
import random

skills = (
    "python",
    "golang",
    "linux",
    "физика", 
    "напряжение",
    "поиск и оптимизация",
    "ускорение",
    "давка",
    "линукс",
    "Деда",
    "макбук",
    "про макс",
    "про навыки",
    "про дададад",
    "про клавиатуру",
    "монитор",
    "проруучка",
)

count = int(input("Count: "))
url = "http://127.0.0.1:8000/api/v1/books?" + "&".join(["text="+random.choice(skills) for i in range(count)])
print(url)