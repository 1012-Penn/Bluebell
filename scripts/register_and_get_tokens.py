import requests
import time

signup_url = "http://127.0.0.1:8084/api/v1/signup"
login_url = "http://127.0.0.1:8084/api/v1/login"
user_prefix = "user"
password = "user1234"
tokens_file = "tokens.txt"

headers = {"Content-Type": "application/json"}
tokens = []

for i in range(1, 10001):
    username = f"{user_prefix}{i:05d}"
    # 注册
    signup_data = {
        "username": username,
        "password": password,
        "re_password": password,
        "confirm_password": password
    }
    try:
        r = requests.post(signup_url, json=signup_data, headers=headers, timeout=5)
        # 注册接口可能已存在用户，忽略错误
    except Exception as e:
        print(f"注册失败: {username}, 错误: {e}")
        continue

    # 登录
    login_data = {
        "username": username,
        "password": password
    }
    try:
        r = requests.post(login_url, json=login_data, headers=headers, timeout=5)
        resp = r.json()
        if "data" in resp and "token" in resp["data"]:
            token = resp["data"]["token"]
            tokens.append(token)
            print(f"{username} 获取token成功")
        else:
            print(f"{username} 登录失败，返回: {resp}")
    except Exception as e:
        print(f"{username} 登录异常: {e}")

    # 可适当sleep，避免接口压力过大
    if i % 100 == 0:
        time.sleep(1)

# 保存到文件
with open(tokens_file, "w") as f:
    for token in tokens:
        f.write(token + "\n")

print(f"共获取到 {len(tokens)} 个token，已保存到 {tokens_file}")