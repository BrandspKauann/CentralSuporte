# Registro Hirayama pela API

Os registros de trabalho usam diretamente `https://hirayama-suporte.vercel.app/api/v2/`.
Não precisam de navegador nem de novo deploy. A API existente já oferece tarefas,
comentários, etiquetas, responsáveis e movimentação no Kanban.

## Conexão neste computador

Com um token de API aprovado da conta KauannPires, execute no PowerShell:

```powershell
.\tools\hirayama\Connect-Hirayama.ps1
```

A entrada fica oculta. O token é validado e salvo fora do repositório, em
`%LOCALAPPDATA%\HirayamaSuporte\api-token.dpapi`, criptografado pela proteção
da conta Windows. Não cole tokens em tarefas, comentários ou commits.
O token precisa das permissões das operações usadas: ler usuário, projetos,
visualizações, colunas, etiquetas e tarefas; criar/atualizar tarefas e comentários;
associar responsáveis e etiquetas; mover tarefas entre colunas. Não conceda
administração, alteração de credenciais nem exclusão para registrar trabalhos.

Exemplos de leitura (a resposta da API é a referência para IDs):

```powershell
.\tools\hirayama\Invoke-HirayamaApi.ps1 -Path user
.\tools\hirayama\Invoke-HirayamaApi.ps1 -Path projects/2
.\tools\hirayama\Invoke-HirayamaApi.ps1 -Path labels
.\tools\hirayama\Invoke-HirayamaApi.ps1 -Path projects/2/views/12/buckets
```

Para gravar, forneça o método e um arquivo JSON sem segredos usando `-BodyFile`.
Consulte o contrato v2 antes de montar o JSON. Após gravar, leia o recurso e
confira responsável, etiquetas, comentários e estado. Não repita automaticamente
um POST quando a resposta falhar: consulte antes para evitar duplicatas.

## Rotina de trabalho

- Procure a tarefa correspondente antes de criar outra no projeto Sites Hirayama.
- Escreva em português e primeira pessoa, como diário de Kauann Pires, relatando
  somente ações efetivamente realizadas. Preserve a autoria real da API.
- Reutilize as etiquetas dos sites e os estados existentes, confirmados pela API.
- Cada solicitação e rodada independente tem tarefa identificável; registre as
  análises, verificações, evidências e próximos passos em comentários datados.
- Só conclua após validar o resultado e a publicação quando aplicável. Sincronize
  `done=true` e a coluna Concluídas, e confira por leitura.
- Sem credencial ou API disponível, mantenha registro local pendente em
  `C:/Users/Administrador/.codex/hirayama/pending`. Não declare gravação remota.

As instruções pessoais e o último mapa da estrutura ficam em
`C:/Users/Administrador/.codex/hirayama/WORKLOG.md` e `structure.json`.
Esses arquivos locais não substituem a confirmação da estrutura pela API.
