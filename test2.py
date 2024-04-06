# Часть 2: Security code review: Python

# Пример №2.2

from flask import Flask, request
import subprocess

app = Flask(name)

@app.route("/dns")
def dns_lookup():
    hostname = request.values.get('hostname')
    cmd = 'nslookup ' + hostname
    output = subprocess.check_output(cmd, shell=True, text=True)
    """
	В нескольких строках выше имеется серьезная уязвимость, связанная с возможностью выполнения произвольных команд
	операционной системы (Command Injection). Уязвимость связана с тем, что переменная hostname, получаемая из пользовательского ввода,
	напрямую конкатенируется с командой nslookup. Это позволяет злоумышленнику вставить произвольные команды, которые будут выполнены на сервере.
	Использование этой уязвимости может привести к удаленному выполнению кода, утечке конфиденциальной информации, компрометации сервера
	или даже полному контролю над системой.
	Для исправления этой уязвимости нужно изменить способ выполнения команды nslookup, к примеру использовать subprocess.run с передачей аргументов
	в виде списка, чтобы предотвратить интерпретацию дополнительных команд shell.
	"""
    return output
if name == "main":
    app.run(debug=True)
    """
	В продакшене необходимо исползовать debug=False иначе уязвимость может привести к утечке чувствительной информации
	и предоставить злоумышленникам дополнительные возможности для атак.
	"""

# С учетом указанных выше потенциальных уязвимостей и проблем в коде, код можно изменить и представить в таком виде

from flask import Flask, request
import subprocess

app = Flask(__name__)

@app.route("/dns")
def dns_lookup():
    hostname = request.values.get('hostname')
    if not hostname:
        return "Hostname is required", 400
    # Валидация hostname может быть более строгой, в зависимости от требований
    if not is_valid_hostname(hostname):
        return "Invalid hostname", 400
    try:
        # Используем безопасный способ вызова команды
        output = subprocess.check_output(["nslookup", hostname], text=True)
    except subprocess.CalledProcessError as e:
        return str(e), 500
    return output

def is_valid_hostname(hostname):
    # Простая проверка на валидность доменного имени
    import re
    if re.match(r'^[a-zA-Z0-9.-]+$', hostname):
        return True
    return False

if __name__ == "__main__":
    app.run(debug=False)  # Убедитесь, что режим отладки отключен в продакшене