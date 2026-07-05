# Modelo de Segurança

[Read in English](./SECURITY_MODEL.md)

## Propósito

Este documento define o modelo de segurança do Gem Bridge.

O Gem Bridge é um daemon local-first que expõe ferramentas controladas de desenvolvimento para assistentes de IA executados no navegador. Seu principal objetivo de segurança é permitir automação local útil sem dar ao assistente acesso irrestrito à máquina do usuário.

O projeto deve tratar toda requisição vinda de um cliente de IA, extensão de navegador, script ou transporte futuro como entrada não confiável.

## Princípio Central de Segurança

A raiz do workspace configurado é a principal fronteira de segurança.

O Gem Bridge só pode operar dentro do workspace autorizado. Qualquer requisição que tente acessar arquivos, diretórios, comandos ou operações Git fora dessa fronteira deve ser rejeitada antes da execução da operação.

Essa regra deve valer igualmente para todos os transportes atuais e futuros:

- CLI
- WebSocket
- Extensão de navegador
- Native Messaging
- Qualquer protocolo local futuro

As regras de segurança devem ficar em pacotes internos compartilhados, não dentro de uma camada específica de transporte.

## Fronteiras de Confiança

O Gem Bridge possui as seguintes fronteiras de confiança:

1. **Fronteira do cliente de IA**
   - Requisições vindas de um assistente de IA não são confiáveis.
   - Nomes de ferramentas, paths, argumentos e parâmetros de comando devem ser validados.

2. **Fronteira de transporte**
   - CLI, WebSocket e Native Messaging são apenas mecanismos de entrega.
   - Um transporte confiável não torna a requisição confiável automaticamente.

3. **Fronteira do workspace**
   - A raiz do workspace é a única área autorizada do filesystem.
   - Todos os paths do filesystem devem ser resolvidos pela camada de segurança.

4. **Fronteira de comandos**
   - Execução de comandos locais é perigosa.
   - O Gem Bridge nunca deve expor acesso irrestrito ao shell.

5. **Fronteira do Git**
   - Operações Git podem modificar histórico, branches e repositórios remotos.
   - Ferramentas Git somente leitura são mais seguras do que operações de escrita.
   - Operações Git sensíveis exigem validação adicional e podem exigir aprovação do usuário no futuro.

## Regras de Segurança para Filesystem

Todos os paths fornecidos pelo usuário devem ser paths relativos.

O Gem Bridge deve rejeitar:

- Paths vazios
- Paths absolutos
- Tentativas de path traversal
- Paths que escapem do workspace através de symlinks
- Paths absolutos específicos de plataforma
- Paths UNC no Windows
- Qualquer path que não possa ser resolvido com segurança dentro do workspace

Exemplos de paths bloqueados:

```text
/etc/passwd
/home/user/.ssh/id_rsa
../../.env
../../.config
C:\Users\user\.ssh\id_rsa
C:/Users/user/.ssh/id_rsa
\\server\share\secret.txt
```

Paths permitidos devem resolver para um local dentro do workspace autorizado depois de limpeza, junção e avaliação de symlinks.

## Segurança com Symlinks

Symlinks são sensíveis para segurança porque um path pode parecer estar dentro do workspace, mas apontar para fora dele.

O Gem Bridge deve bloquear escapes por symlink validando paths existentes e paths pais antes de executar operações no filesystem.

Exemplo de estrutura perigosa:

```text
workspace/
  link-to-home -> /home/user
```

Uma requisição como esta deve ser rejeitada:

```json
{
  "tool": "readFile",
  "path": "link-to-home/.ssh/id_rsa"
}
```

A camada de segurança deve garantir que o path final resolvido continue dentro da raiz do workspace.

## Operações de Leitura

Operações de leitura são a capacidade inicial mais segura do filesystem, mas ainda são sensíveis.

Ferramentas de leitura devem:

- Exigir path relativo
- Resolver o path pela camada de segurança do workspace
- Recusar acesso fora do workspace
- Retornar erros em JSON estruturado
- Evitar vazar detalhes desnecessários do filesystem do host quando possível

Ferramentas atuais orientadas à leitura:

- `listDirectory`
- `readFile`

Ferramentas futuras de leitura devem seguir o mesmo caminho de segurança antes de tocar no disco.

## Operações de Escrita

Operações de escrita são mais perigosas do que operações de leitura e devem ser implementadas com regras mais rígidas.

Ferramentas futuras de escrita devem:

- Exigir paths relativos
- Resolver paths pela camada de segurança do workspace
- Bloquear traversal e escapes por symlink
- Recusar escrita fora do workspace
- Evitar sobrescrever arquivos existentes, exceto quando isso for explicitamente permitido
- Considerar limites de tamanho de arquivo
- Retornar erros em JSON estruturado
- Ter cobertura por testes automatizados
- Futuramente oferecer aprovação do usuário para escritas sensíveis

A escrita inicial de arquivos deve ser estreita e explícita. Devemos evitar ferramentas amplas que possam modificar arquivos arbitrários sem regras claras.

## Regras de Execução de Comandos

O Gem Bridge não deve expor execução arbitrária de shell.

O daemon não deve aceitar requisições como:

```json
{
  "tool": "runCommand",
  "command": "rm -rf ~/project"
}
```

Em vez disso, comportamentos semelhantes a comandos devem ser expostos por ferramentas explícitas com argumentos controlados.

Exemplos mais seguros:

- `goFmt`
- `goTest`
- `gitStatus`
- `gitDiff`

Cada ferramenta de comando deve:

- Chamar um executável conhecido
- Usar argumentos controlados
- Rodar dentro do workspace autorizado
- Evitar interpolação de shell
- Evitar comportamento destrutivo por padrão
- Retornar saída estruturada
- Ser testada quando possível

Comandos perigosos e recursos de shell devem permanecer bloqueados, exceto se um modelo futuro de aprovação oferecer suporte explícito.

## Regras para Operações Git

Ferramentas Git devem ser introduzidas de forma incremental.

Ferramentas Git iniciais mais seguras:

- `gitStatus`
- `gitDiff`

Elas são orientadas à leitura e úteis para validação.

Ferramentas Git mais sensíveis:

- `gitAdd`
- `gitCommit`
- `gitCheckout`
- `gitMerge`
- `gitPush`

Ferramentas Git sensíveis podem alterar histórico, branches, estado remoto ou conteúdo staged. Elas devem exigir validação mais rígida e podem exigir aprovação do usuário em uma versão futura.

Ferramentas Git devem rodar apenas dentro do workspace autorizado.

## Fluxo de Aprovação

O Gem Bridge ainda não implementa um fluxo de aprovação, mas a arquitetura deve se preparar para isso.

Operações futuras que podem exigir aprovação incluem:

- Escrever arquivos
- Sobrescrever arquivos
- Deletar arquivos
- Rodar testes ou comandos de build com efeitos colaterais
- Adicionar arquivos ao stage
- Criar commits
- Trocar branches
- Fazer merge de branches
- Fazer push para remotos
- Instalar dependências
- Rodar qualquer ferramenta que possa modificar o sistema local

A aprovação deve ser explícita, auditável e compreensível para o usuário.

## Tratamento de Erros

Falhas de segurança devem retornar erros estruturados.

Uma requisição com falha não deve revelar detalhes desnecessários do host.

No modo CLI, retornar um status diferente de zero para requisições de ferramenta com falha é aceitável.

No futuro modo daemon, uma falha de ferramenta não deve derrubar o processo. O daemon deve retornar um erro JSON estruturado e continuar em execução.

## Considerações de Segurança Cross-Platform

A segurança de paths deve ser validada em Linux, macOS e Windows.

É preciso cuidado especial com:

- Paths absolutos Unix
- Paths Windows com letra de unidade
- Paths UNC no Windows
- Separadores de path misturados
- Comportamento de symlinks e junctions
- Diferenças de case sensitivity
- Localização de diretórios de configuração
- Paths de registro de host Native Messaging

A CI cross-platform deve futuramente executar testes de segurança em:

```text
ubuntu-latest
macos-latest
windows-latest
```

## Não Objetivos

O Gem Bridge não deve oferecer suporte aos seguintes comportamentos no estágio atual:

- Execução arbitrária de shell
- Acesso irrestrito ao filesystem
- Exposição remota do daemon
- Automação de GUI
- Operações destrutivas silenciosas
- Operações Git sensíveis sem validação
- Ferramentas de escrita sem regras claras de segurança
- Atalhos específicos de plataforma que ignorem a camada de segurança compartilhada

## Requisitos de Testes

Comportamentos de segurança devem ser testados antes que funcionalidades sejam consideradas completas.

Testes atuais e futuros devem cobrir:

- Criação da raiz do workspace
- Resolução de paths relativos
- Rejeição de path vazio
- Rejeição de path absoluto
- Rejeição de path traversal
- Rejeição de escape por symlink
- Leitura dentro do workspace
- Leitura fora do workspace
- Futuras regras de escrita segura
- Futuro comportamento de allowlist de comandos
- Futura validação de ferramentas Git

Testes de segurança devem ser pequenos, explícitos e fáceis de entender.

## Regra Final

Na dúvida, o Gem Bridge deve escolher o comportamento mais seguro.

Uma requisição segura bloqueada é um inconveniente. Uma requisição insegura permitida pode expor arquivos privados, danificar um projeto ou executar ações locais indesejadas.
