# Modelo de Segurança

[Read in English](./SECURITY_MODEL.md)

## Propósito

Este documento define o modelo de segurança do Gem Bridge.

O Gem Bridge é um daemon local-first que expõe ferramentas controladas de desenvolvimento para assistentes de IA executados no navegador. Seu principal objetivo de segurança é permitir automação local útil sem dar ao assistente acesso irrestrito à máquina do usuário.

O projeto deve tratar toda requisição vinda de um cliente de IA, extensão de navegador, script ou transporte futuro como entrada não confiável.

## Princípio Central de Segurança

A raiz do workspace configurado é a principal fronteira de segurança.

O Gem Bridge só pode operar dentro do workspace autorizado. Qualquer requisição que tente acessar arquivos, diretórios, comandos ou operações Git fora dessa fronteira deve ser rejeitada antes da execução da operação.

As regras de segurança devem ficar em pacotes internos compartilhados, não dentro de uma camada específica de transporte.

## Fronteiras de Confiança

1. **Fronteira do cliente de IA**
   - Requisições vindas de um assistente de IA não são confiáveis.
   - Nomes de ferramentas, paths, argumentos, payloads de conteúdo e parâmetros de comando devem ser validados.

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

Paths permitidos devem resolver para um local dentro do workspace autorizado depois de limpeza, junção, normalização de separadores e avaliação de symlinks.

## Segurança com Symlinks

Symlinks são sensíveis para segurança porque um path pode parecer estar dentro do workspace, mas apontar para fora dele.

O Gem Bridge deve bloquear escapes por symlink validando paths existentes e paths pais antes de executar operações no filesystem.

## Operações de Leitura

Ferramentas atuais orientadas à leitura:

- `listDirectory`
- `readFile`

Ferramentas de leitura devem exigir path relativo, resolver o path pela camada de segurança do workspace, recusar acesso fora do workspace e retornar erros em JSON estruturado.

## Operações de Escrita

Ferramentas atuais orientadas à escrita:

- `writeFile`

A implementação atual de `writeFile` é intencionalmente conservadora. Ela:

- Exige path relativo
- Resolve o path pela camada de segurança do workspace
- Bloqueia traversal e escapes por symlink
- Recusa escrita fora do workspace
- Recusa sobrescrever arquivos existentes
- Aplica limite máximo de tamanho do conteúdo
- Retorna erros em JSON estruturado
- É coberta por testes automatizados

Esta primeira versão cria apenas arquivos de texto novos. Sobrescrever, aplicar patches, anexar conteúdo, deletar ou renomear arquivos devem permanecer como ferramentas futuras separadas, com regras explícitas de segurança.

Capacidades futuras de escrita devem evitar ferramentas amplas que possam modificar arquivos arbitrários sem regras claras.

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

Cada ferramenta de comando deve chamar um executável conhecido, usar argumentos controlados, rodar dentro do workspace autorizado, evitar interpolação de shell, evitar comportamento destrutivo por padrão, retornar saída estruturada e ser testada quando possível.

## Regras para Operações Git

Ferramentas Git devem ser introduzidas de forma incremental.

Ferramentas Git iniciais mais seguras:

- `gitStatus`
- `gitDiff`

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

Operações futuras que podem exigir aprovação incluem sobrescrever arquivos, deletar arquivos, rodar comandos com efeitos colaterais, adicionar arquivos ao stage, criar commits, trocar branches, fazer merge, fazer push, instalar dependências ou rodar ferramentas mais amplas que modifiquem o sistema.

A aprovação deve ser explícita, auditável e compreensível para o usuário.

## Tratamento de Erros

Falhas de segurança devem retornar erros estruturados.

No modo CLI, retornar um status diferente de zero para requisições de ferramenta com falha é aceitável.

No futuro modo daemon, uma falha de ferramenta não deve derrubar o processo. O daemon deve retornar um erro JSON estruturado e continuar em execução.

## Considerações de Segurança Cross-Platform

A segurança de paths deve ser validada em Linux, macOS e Windows.

É preciso cuidado especial com paths absolutos Unix, paths Windows com letra de unidade, paths UNC no Windows, separadores de path misturados, comportamento de symlinks e junctions, diferenças de case sensitivity, localização de diretórios de configuração e paths de registro de host Native Messaging.

A CI cross-platform atualmente executa formatação e testes Go em:

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
- Ferramentas amplas de escrita sem regras claras de segurança
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
- Criação de arquivo dentro do workspace
- Rejeição de sobrescrita
- Rejeição por limite de tamanho
- Rejeição de escrita em paths inseguros
- Rejeição de escrita através de symlink em diretório pai
- Futuro comportamento de allowlist de comandos
- Futura validação de ferramentas Git

## Regra Final

Na dúvida, o Gem Bridge deve escolher o comportamento mais seguro.

Uma requisição segura bloqueada é um inconveniente. Uma requisição insegura permitida pode expor arquivos privados, danificar um projeto ou executar ações locais indesejadas.
