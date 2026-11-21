# Avisos de conclusão por e-mail

A fila persistente `completion_mail` acompanha um envio por tarefa e destinatário. Apenas o projeto configurado e conclusões a partir da data de ativação são elegíveis. Tarefas antigas não são reenviadas; reabrir e concluir novamente não gera outro aviso. Comentários, descrições e anexos não entram no e-mail: envio o título e o link autenticado da tarefa.

Configuração exclusivamente no servidor:

- `HIRAYAMA_COMPLETION_EMAIL_ENABLED=true`
- `HIRAYAMA_COMPLETION_EMAIL_PROJECT_ID=2`
- `HIRAYAMA_COMPLETION_EMAIL_ENABLED_AFTER`: data UTC da ativação em RFC3339.
- `HIRAYAMA_COMPLETION_EMAIL_TO`: destinatários autorizados separados por vírgula.
- `HIRAYAMA_COMPLETION_SMTP_HOST`, `HIRAYAMA_COMPLETION_SMTP_PORT`, `HIRAYAMA_COMPLETION_SMTP_USERNAME`, `HIRAYAMA_COMPLETION_SMTP_PASSWORD` e `HIRAYAMA_COMPLETION_SMTP_FROM`. O transporte é separado do mailer geral: não ativa lembretes, cobranças ou e-mails de conta.

A senha deve ser variável sensível, somente em produção. Nunca usar prefixo público. O SMTP utiliza certificado TLS validado. A autenticação já foi conferida sem envio de mensagem em 21/11/2026.

A conclusão e a fila são gravadas na mesma transação, tanto pelo formulário da tarefa quanto pelo quadro Kanban. O processamento acontece após o evento e a cada minuto; cada registro usa uma reserva de dois minutos para impedir processamento concorrente. Falhas recebem até cinco tentativas com intervalos crescentes. O estado `sent` indica aceitação pelo servidor SMTP, sem garantir leitura ou caixa de entrada. Falhas definitivas ficam em `failed`, visíveis na tabela e nos logs por ID da tarefa, sem credenciais ou conteúdo SMTP.

SMTP não oferece entrega exatamente uma vez: se o servidor aceitar o e-mail e a conexão ou gravação cair antes da confirmação, uma repetição ainda é possível. Duplicações normais de evento, edição posterior e reinício são bloqueadas pela chave única e pelo estado persistente.

Para suspender, definir `HIRAYAMA_COMPLETION_EMAIL_ENABLED=false` e republicar. A fila permanece preservada. Não reativar com destinatários de outra empresa: o escopo é exclusivamente Sites Hirayama.

Validação: `mage test:completionMail`; a suíte usa remetente simulado e nunca envia e-mails reais a clientes.
