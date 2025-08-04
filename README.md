# 🚀 Projeto de Fechamento Automático de Leilão

## 📌 Objetivo

Este projeto tem como objetivo adicionar uma nova funcionalidade ao sistema de leilões: **fechamento automático do leilão após um tempo determinado via variável de ambiente**.


## 🛠️ Como rodar o projeto em ambiente de desenvolvimento

### 1. Clone o repositório

```
git clone https://github.com/afga95/labs-auction.git
cd labs-auction

```

### 2. Subir container

Como as variáveis de ambiente (.env) já estão pré definidas pelo código original, não é necessário criar o arquivo.

```
docker-compose up -d --build

```

```
docker-compose down -v

```