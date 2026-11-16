# Hirayama Suporte

Plataforma de projetos e atendimento baseada no **Vikunja Community v2.8.0**.
Interface com a lhama original e a marca Hirayama Suporte. A senha é preservada
quando o usuário alterna sua visibilidade no cadastro, redefinição e troca.

## Implantação completa na Vercel

O arquivo `vercel.json` define dois serviços no mesmo projeto:

| Serviço | Execução | Caminho público |
|---|---|---|
| `frontend` | Vue/Vite | `/` e rotas da interface |
| `app` | API Vikunja em container | `/api/*` |

As rotas `/api/v1/*` e `/api/v2/*` chegam à API sem remover o prefixo.
A interface usa a mesma origem e preserva os cookies de sessão. Não há bindings
internos: o Vite gera arquivos estáticos, e quem chama `/api` é o navegador.
`VITE_API_URL` não é necessária nessa implantação e pode ser removida da Vercel.

1. Importe `BrandspKauann/Hirayama-Suporte` na Vercel, com raiz `./` e Node 24.x.
2. Use a configuração de **Services** fornecida pelo repositório.
3. Crie um PostgreSQL gerenciado pelo Marketplace da Vercel. A integração Supabase
   também oferece armazenamento compatível com S3 para os anexos.
4. Conecte o recurso ao projeto. O inicializador reconhece `POSTGRES_URL_NON_POOLING`,
   `DATABASE_URL` ou `POSTGRES_URL` e converte para a configuração do Vikunja.
   A conexão direta tem preferência para permitir as migrações do banco.
5. Crie um bucket **privado** para anexos e obtenha suas credenciais S3.
6. Configure as variáveis abaixo e execute Deploy.

### Variáveis de ambiente

Defina no projeto Vercel; nunca coloque os valores secretos no GitHub.
O modelo completo está em `deploy/vercel-api/.env.example`.

| Variável | Valor |
|---|---|
| `DATABASE_URL` | Conexão PostgreSQL com TLS, se não for fornecida pela integração |
| `VIKUNJA_SERVICE_SECRET` | Chave aleatória fixa, com pelo menos 64 caracteres |
| `VIKUNJA_FILES_S3_ENDPOINT` | Endpoint HTTPS do armazenamento |
| `VIKUNJA_FILES_S3_BUCKET` | Nome do bucket privado |
| `VIKUNJA_FILES_S3_REGION` | Região do bucket |
| `VIKUNJA_FILES_S3_ACCESSKEY` | Chave de acesso S3 |
| `VIKUNJA_FILES_S3_SECRETKEY` | Chave secreta S3 |
| `VIKUNJA_FILES_S3_USEPATHSTYLE` | `true` para Supabase Storage |
| `VIKUNJA_SERVICE_PUBLICURL` | Opcional: URL HTTPS da interface; padrão `https://$VERCEL_URL/` |
| `VIKUNJA_SERVICE_ENABLEREGISTRATION` | `false`; habilite temporariamente para criar a conta inicial |

Gere a chave fixa com `node -e "console.log(require('crypto').randomBytes(32).toString('hex'))"`.
A mesma chave precisa ser mantida nos próximos deployments para preservar sessões.

No Supabase, o endpoint S3 tem o formato
`https://PROJECT_REF.supabase.co/storage/v1/s3`. Crie o bucket `hirayama-anexos`
como privado, gere credenciais S3 no painel e use a região indicada pelo provedor.
As chaves S3 não são as chaves públicas/secretas da API Supabase.

Não configure PostgreSQL, SQLite, MinIO ou anexos como serviços com disco local
na Vercel: a persistência fica nos provedores gerenciados. A API exige PostgreSQL
com TLS e S3 em produção; recusa iniciar com dados persistentes em disco temporário.
Arquivos temporários, cache e logs podem usar `/tmp` e stdout.

### O que foi validado e limitações

O container usa o binário oficial v2.8.0, pois ainda não existem modificações no
backend Go. As personalizações atuais ficam no frontend, compilado separadamente.
O inicializador adapta `PORT`, conexão do banco, HTTPS e armazenamento persistente,
sem mudar permissões, modelos ou o sistema de licença.

A Vercel pode suspender instâncias sem tráfego. Os agendamentos internos do
Vikunja não têm garantia de execução nesse período. WebSockets também precisam
ser avaliados conforme os limites da plataforma; atualização em tempo real e
lembretes agendados não estão validados nesta implantação. Projetos, tarefas,
login e anexos usam as rotas HTTP normais. E-mails precisam de SMTP configurado.

Uma implantação nova não migra os dados locais automaticamente. Faça backup e
importe o PostgreSQL e os anexos antes de substituir uma instalação existente.

## Desenvolvimento e validação

Node 24, pnpm 12.16.1 e Docker disponíveis:

```sh
cd frontend
pnpm install --frozen-lockfile
pnpm dev
```

Para testar os dois serviços juntos, configure as variáveis de armazenamento e
execute `npx vercel dev --local` na raiz. `PORT` é fornecida pela Vercel e não deve
ser fixada manualmente. As variáveis secretas de produção não entram em `.env`
versionado nem no bundle da interface.

Para validar o inicializador: `node --test deploy/vercel-api/config.test.mjs`.
O CI verifica esse inicializador, a correção da senha e a compilação da interface.
A configuração anterior com API em servidor Docker externo continua disponível
como alternativa em `deploy/docker`, mas não participa do deploy Vercel Services.

## Origem e licença

Código original: https://github.com/go-vikunja/vikunja (tag `v2.8.0`).
O histórico, a licença AGPL-3.0-or-later e os avisos originais foram preservados.
Consulte `LICENSE` e `README.upstream.md`. O código das modificações está neste
repositório público. Fluxos originais de publicação estão arquivados em
`.github/upstream-workflows`.

Documentação: https://vercel.com/docs/services,
https://vercel.com/docs/functions/runtimes/container,
https://vercel.com/docs/marketplace-storage.
