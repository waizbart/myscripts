Você tem acesso ao meu código. Antes de gerar qualquer coisa, leia os arquivos relevantes para entender a stack, convenções e padrões já existentes no projeto.

Quero criar um prd.json para usar com o Ralph — um loop autônomo de agente que executa uma story por vez, em instâncias separadas com contexto limpo. Cada story precisa ser pequena o suficiente para caber em uma única context window.

Me faça as seguintes perguntas antes de gerar o prd.json:
1. Qual é a feature que quero implementar?
2. É uma nova feature, um fix ou um refactor?
3. Há alguma restrição ou convenção específica que devo seguir?
4. Tenho testes? Se sim, qual comando roda?

Com base nas minhas respostas e no que você leu do código, gere um prd.json seguindo estas regras:

REGRAS PARA AS STORIES:
- Cada story deve ser atômica — uma única responsabilidade
- A descrição deve conter: o que fazer, onde fazer (arquivo/pasta), e como verificar que está correto
- Stories com dependência entre si devem estar em ordem no array
- Nunca inclua "implementar X inteiro" — quebre sempre em partes menores
- O critério de conclusão deve ser verificável por comando (tsc, lint, test, curl, etc.)

REGRAS DE REAPROVEITAMENTO E MODULARIZAÇÃO:
- Antes de criar qualquer arquivo ou função, inspecione o código existente para identificar utilitários, hooks, helpers, componentes ou serviços que possam ser reutilizados
- Se existir algo reutilizável, a story deve explicitamente referenciar onde está e como usar — nunca duplicar
- Se a implementação exigir lógica genérica, a story deve orientar a extraí-la como módulo reutilizável desde o início
- Nunca escreva uma função que já existe em outro lugar do projeto — verificar antes

REGRAS PARA FIXES (quando for tipo "fix"):
- A primeira story do fix deve ser um teste de regressão: escrever um teste unit que reproduz o bug e falha ANTES do fix
- Se o bug afeta um fluxo de usuário, incluir também um teste E2E que reproduz o comportamento quebrado
- Só depois escrever a story do fix em si — que deve fazer os testes de regressão passarem
- Marcar a story com "[regression]" no título quando for exclusivamente para adicionar testes de regressão

REGRAS PARA BREAKING CHANGES:
- Se uma story altera contrato público (API, schema de banco, tipos exportados, eventos), sinalize com "[breaking]" no título
- A descrição deve listar o que quebra e o que precisa ser atualizado em cascata (migrations down, consumidores, documentação)

REGRA DE CLEANUP:
- Ao final do array, se a implementação provavelmente deixará código morto (funções antigas, imports não usados, flags removidas), adicione uma story de cleanup explícita
- Exemplo: "Remover função legada X de src/utils/old.ts após migração para Y. Verificar com: npx tsc --noEmit e grep -r 'oldFunction' retornando vazio."

EXEMPLOS DE STORIES BEM ESCRITAS:
- "Adicionar coluna expires_at na tabela sessions via migration. Verificar com: npx tsc --noEmit passando sem erros."
- "Criar resolver GraphQL getUserById em src/resolvers/user.ts seguindo o padrão dos resolvers existentes. Verificar com: npx tsc --noEmit."

EXEMPLOS DE STORIES MAL ESCRITAS (nunca faça assim):
- "Implementar autenticação"
- "Criar o módulo de usuários"
- "Refatorar o código"

O prd.json deve seguir exatamente este formato, sem propriedades a mais ou a menos:
{
  "branchName": "feature/nome-kebab-case",
  "userStories": [
    {
      "id": "1",
      "title": "Título curto e específico",
      "description": "O que fazer, onde, e como verificar que está correto.",
      "passes": false
    }
  ]
}