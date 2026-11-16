# Hirayama Suporte

Plataforma de projetos e atendimento baseada no **Vikunja Community v2.8.0**.
Mantém a lhama original, com a marca Hirayama Suporte. Inclui a correção que
preserva a senha ao alternar sua visibilidade no cadastro, redefinição e troca.

## Publicar a interface na Vercel

1. Importe `BrandspKauann/Hirayama-Suporte` na Vercel.
2. Mantenha **Root Directory** na raiz do repositório (`./`).
3. Em **Node.js Version**, selecione **24.x**. O `vercel.mjs` configura instalação,
   compilação, saída `frontend/dist` e rotas da aplicação.
4. Cadastre a variável de ambiente **VITE_API_URL** com a raiz HTTPS pública da API
   Docker, por exemplo `https://api.seudominio.com`, **sem `/api/v2`**.
   Configure em Production e Preview se quiser usar a mesma API nos dois ambientes.
5. Clique em Deploy. Se a variável estiver ausente ou apontar para localhost, a
   compilação explica o que falta, em vez de publicar um login sem conexão.

`VITE_API_URL` é público. Senhas do banco, chaves e tokens nunca entram na Vercel
nem no código da interface. Uma mudança nessa variável exige novo deploy.

## API e PostgreSQL em Docker

A interface usa `/api` na Vercel, que encaminha as requisições à API Docker. Isso mantém os cookies de sessão na mesma origem e permite usar o domínio padrão vercel.app. O banco e os anexos
usam volumes persistentes. O banco não expõe porta pública.

1. Use um servidor acessível pela internet, com Docker Compose instalado.
2. Aponte o DNS de um subdomínio (ex.: `api.seudominio.com`) para o IP do servidor.
   Libere as portas 80 e 443 para o Caddy emitir o certificado HTTPS.
3. Entre em `deploy/docker` e copie `.env.example` para `.env`.
4. Defina domínio, e-mail, origem exata da Vercel e duas chaves aleatórias fortes.
   Uma forma de gerar cada chave: `openssl rand -hex 32`.
5. Execute `docker compose up -d` dentro dessa pasta.
6. Confira `https://api.seudominio.com/api/v2/info` e depois o login na Vercel.

Para criar a conta inicial, habilite `ENABLE_REGISTRATION=true`, execute
`docker compose up -d`, crie a conta pela interface e volte para `false`, executando
`docker compose up -d` novamente. Se adicionar um domínio próprio à interface,
atualize `FRONTEND_ORIGIN` e reinicie o serviço `api`. Previews da Vercel têm origens
próprias; só permita as que forem usadas, sem liberar todos os domínios.

**Docker no seu computador não cria um endereço público automaticamente.** Para
usar esse computador como servidor é necessário configurar acesso externo e DNS;
o endereço `localhost:3456` não funciona para visitantes da Vercel.

A implantação cria um banco novo. Os dados da instalação local não são enviados
para o GitHub nem migrados automaticamente. Faça backup do PostgreSQL e do volume
de anexos antes de migrar os dados. E-mails de redefinição de senha precisam de SMTP
configurado no backend; não estão configurados neste pacote.

## Desenvolvimento

Requer Node 24 e pnpm 12.16.1. Dentro de `frontend`:

```sh
pnpm install --frozen-lockfile
pnpm dev
```

Sem `VITE_API_URL`, a interface usa a API da mesma origem, como a instalação local.
Para usar uma API separada em desenvolvimento, copie `frontend/.env.example` para
`frontend/.env.local` e ajuste a URL. Configure também a origem local no CORS da API.

## Origem e licença

Código-fonte original: https://github.com/go-vikunja/vikunja (tag `v2.8.0`).
O histórico, a licença AGPL-3.0-or-later e os avisos originais foram preservados.
Consulte `LICENSE` e `README.upstream.md`. O código correspondente às modificações
está neste repositório público. Os fluxos de publicação do projeto original foram
arquivados em `.github/upstream-workflows`; o CI deste projeto só valida a interface.

Documentação oficial da configuração: https://vercel.com/docs/project-configuration/vercel-ts e https://vercel.com/docs/routing/rewrites.
