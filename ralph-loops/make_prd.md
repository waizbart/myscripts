Você tem acesso ao meu código. Antes de gerar qualquer coisa, leia os arquivos relevantes para entender a stack, convenções e padrões já existentes no projeto.

Quero criar um prd.json para usar com o Ralph — um loop autônomo de agente que executa uma story por vez, em instâncias separadas com contexto limpo. Cada story precisa ser pequena o suficiente para caber em uma única context window.

Me faça as seguintes perguntas antes de gerar o prd.json:
1. Qual é a feature que quero implementar?
2. Há alguma restrição ou convenção específica que devo seguir?
3. Tenho testes? Se sim, qual comando roda?

Com base nas minhas respostas e no que você leu do código, gere um prd.json seguindo estas regras:

REGRAS PARA AS STORIES:
- Cada story deve ser atômica — uma única responsabilidade
- A descrição deve conter: o que fazer, onde fazer (arquivo/pasta), e como verificar que está correto
- Stories com dependência entre si devem estar em ordem no array
- Nunca inclua "implementar X inteiro" — quebre sempre em partes menores
- O critério de conclusão deve ser verificável por comando (tsc, lint, test, curl, etc.)

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