# Часть 2: Security code review: Python

# Пример №2.1

from flask import Flask, request
from jinja2 import Template

app = Flask(name)

@app.route("/page")
def page():
    name = request.values.get('name')
    age = request.values.get('age', 'unknown')
    """ 
    В двух строках выше используется прямое включение пользовательских данных (name и age) в шаблон без их предварительной санитизации.
	Это может привести к XSS-атакам. Правильно было бы использовать функцию escape из Flask для санитизации пользовательских данных.
    """
    output = Template('Hello ' + name + '! Your age is ' + age + '.').render()
    """
	Использование Template из jinja2 в таком виде потенциально опасно и может привести к выполнению вредоносного кода в шаблоне. 
	В данном случае рациональнее f-строку для возврата форматированного значения output. 
	"""
    return output

if name == "main":
    app.run(debug=True)
    """
	В продакшене необходимо исползовать debug=False иначе уязвимость может привести к утечке чувствительной информации
	и предоставить злоумышленникам дополнительные возможности для атак.
	"""


# С учетом указанных выше потенциальных уязвимостей и проблем в коде, код можно изменить и представить в таком виде
from flask import Flask, request, escape

app = Flask(__name__)

@app.route("/page")
def page():
    name = escape(request.values.get('name', ''))
    age = escape(request.values.get('age', 'unknown'))
    # Используем безопасный способ вставки данных в шаблон
    output = f"Hello {name}! Your age is {age}."
    return output

if __name__ == "__main__":
    # Убедитесь, что режим отладки отключен в продакшене
    app.run(debug=False)