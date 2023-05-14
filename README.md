# Utilitário em golang para subir instâncias EC2 na AWS.

**A ideia é facilitar a criação de máquinas virtuais para teste na AWS sem precisar abrir o console**

<br>

**Regions:**
- us-east-1
- us-west-1
- sa-east-1

<br>

**Imagens:**
- ubuntu
- awslinux
- redhat

<br>

**Forma de uso:**

Clonar o repositório e executar:
```
go run main.go
```
ou utilizar o binário ec2_launch
```
./ec2_launch
```
<br>

Logo em seguida deverá:

- Escolher a imagem (ubuntu, awslinux ou redhat).
- Escolher a região (us-east-1, us-west-1 ou sa-east-1).
- Escolher o nome da chave SSH(Pode ser uma já criada na AWS. Caso não tenha, ele irá criar a chave na AWS e criar a .pem no seu diretório).
- Você terá um output do ID da instância e do IP para fazer conexão SSH.
- E por final pode escolher se deseja deletar a instância criada.

