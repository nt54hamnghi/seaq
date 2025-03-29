# seaq

`seaq` (pronounced "seek") allows you to extract text data from the web and process it with your favorite prompt and LLM model, all from your terminal.

`seaq` was initially created as an experiment in using LLMs to implement the idea of **optimizing cognitive load**, as presented by Dr. Justin Sung in [this video](https://www.youtube.com/watch?v=1iv4YPQVmTc). It was strongly inspired by [`fabric`](https://github.com/danielmiessler/fabric) and was originally meant to be a Go-based alternative (as it turns out, `fabric` is now written in Go too).

## Features

- Support for multiple LLM providers.
  - Built-in (OpenAI, Anthropic, Google).
  - Any OpenAI-compatible providers via `connection`.
- Scrape web pages with various engines.
  - Built-in
  - [Firecrawl](https://www.firecrawl.dev)
  - [Jina](https://jina.ai)
- Fetch YouTube transcripts, Udemy transcripts, X threads.
- Adding patterns from a GitHub repository on demand.
- YAML-based configuration file.

## Example workflows

![seaq demo](./imgs/seaq.gif)

```sh
# Fetch a YouTube video transcript with defaults in the `seaq.yaml` config file
seaq fetch youtube "446E-r0rXHI" | seaq
```

```sh
# Get insights from an X thread using a local model
seaq fetch x "1883686162709295541" | seaq --pattern prime_mind --model ollama/smollm2:latest
```

```sh
# Fetch a web page and chat with it
# `--auto` tells the built-in scrapper to automatically detect the main content of the page
seaq fetch page "https://modelcontextprotocol.io/introduction" --auto | seaq chat
```

## Installation

[Make sure Go is installed](https://go.dev/doc/install) before running the following command:

```sh
go install github.com/nt54hamnghi/seaq@latest
```

You may need to add the following environment variables to run `seaq`:

```sh
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
# Add Go binary paths and local user binaries to the system PATH
export PATH=$GOPATH/bin:$GOROOT/bin:$HOME/.local/bin:$PATH
```

## Getting started

Run `seaq config setup` to generate a new config file interactively. It will walk you through selecting:

- An LLM model
- A pattern repository (where you store your patterns)
- A pattern (prompt)
- A default remote GitHub repository to download patterns

These values act as defaults and can be overridden by CLI flags and/or arguments.

```sh
seaq config setup
```

Once completed, you will have a minimal config file like this:

```yaml
model:
    name: anthropic/claude-3-5-sonnet-latest
pattern:
    name: take_note
    remote: https://github.com/danielmiessler/fabric
    repo: /home/nt54hamnghi/.config/seaq/patterns
```

> Note:
> `seaq` reads API keys from environment variables. For example, if you're using a model from OpenAI, `seaq` expects the `OPENAI_API_KEY` environment variable to hold the API key.

Supported environment variables:

- `OPENAI_API_KEY`
- `ANTHROPIC_API_KEY`
- `GEMINI_API_KEY`
- `YOUTUBE_API_KEY`
- `CHROMA_URL`
- `X_AUTH_TOKEN`
- `X_CSRF_TOKEN`
- `UDEMY_ACCESS_TOKEN`
- `OLLAMA_HOST`
- `JINA_API_KEY`
- `FIRECRAWL_API_KEY`
- `SEAQ_SUPPRESS_WARNINGS`

## Usage

### `seaq` command

The root command, `seaq`, processes input text with a specified model and pattern. It can read input either from a file or from standard input (pipe).

```sh
# Process input from a file
seaq -i input.txt

# Process piped input
echo "Some text" | seaq
```

Available flags:

```sh
-m, --model string        model to use
    --hint string         optional context to guide the LLM's focus
    --no-stream           disable streaming mode
    --temperature float   temperature to use (default 0.7)
-p, --pattern string      pattern to use
-r, --repo string         path to the pattern repository
-i, --input string        input file
-c, --config string       config file (default is $HOME/.config/seaq.yaml)
-o, --output string       output file
-f, --force               overwrite existing file
```

### Chat with a model

> Note: `seaq chat` is an experimental feature.

To use the chat feature, you need a running [ChromaDB](https://www.trychroma.com/) instance. To quickly start one:

```sh
docker run -d -p 7645:8000 --name chroma-core ghcr.io/chroma-core/chroma:0.5.0

# `seaq` requires the `CHROMA_URL` environment variable to be set
export CHROMA_URL=http://0.0.0.0:7645
```

Like the root command, `seaq chat` accepts input from both a file and standard input (pipe).

```sh
# Process input from a file
seaq -i essay.md

# Process piped input
cat essay.md | seaq chat
```

### Manage patterns and models

```sh
# List all available patterns
seaq pattern list

# Set and get the default pattern
# These will modify the `seaq.yaml` config file
seaq pattern set prime_mind
seaq pattern get
```

```sh
# List all available models
seaq model list

# Set and get the default model
# These will modify the `seaq.yaml` config file
seaq model set ollama/smollm2:latest
seaq model get
```

### Fetch data

`seaq fetch` can fetch data from a variety of sources.

#### 1. Web pages

```sh
seaq fetch page "https://en.wikipedia.org/wiki/Go_(programming_language)"
```

Or you can use `firecrawl` or `jina` as the scraping engine with `--engine` flag.

```sh
seaq fetch page --engine firecrawl "https://en.wikipedia.org/wiki/Go_(programming_language)"
```

```sh
seaq fetch page --engine jina "https://en.wikipedia.org/wiki/Go_(programming_language)"
```

> Note:
>
> - `jina` requires setting the `JINA_API_KEY` environment variable
> - `firecrawl` requires setting the `FIRECRAWL_API_KEY` environment variable

#### 2. YouTube video transcripts

```sh
seaq fetch youtube "https://www.youtube.com/watch?v=446E-r0rXHI"
```

Or just with the video ID:

```sh
seaq fetch youtube "446E-r0rXHI"
```

You can use `--start` and `--end` flags to filter the transcript.

```sh
# Fetch the transcript from 0:07 to 0:42
seaq fetch youtube "446E-r0rXHI" --start 0:07 --end 0:42
```

To include metadata (e.g., video title, description, etc.) in the output, use the `--metadata` flag.

```sh
seaq fetch youtube "446E-r0rXHI" --metadata
```

> **Note:**
> Fetching metadata requires setting the `YOUTUBE_API_KEY` environment variable.

#### 3. Udemy courses transcripts

```sh
seaq fetch udemy "https://www.udemy.com/course/course-name/learn/lecture/lecture-id"
```

`seaq fetch udemy` also supports the `--start` and `--end` flags.

```sh
seaq fetch udemy "https://www.udemy.com/course/course-name/learn/lecture/lecture-id" --start 0:07 --end 0:42
```

> Note:
>
> `seaq fetch udemy` requires setting the `UDEMY_ACCESS_TOKEN` environment variable. You can get this token by inspecting the cookies after logging in to your Udemy account.

#### 4. X tweets and threads

```sh
seaq fetch x "https://x.com/morganb/status/1883686162709295541"
```

Or just with the tweet ID:

```sh
seaq fetch x "1883686162709295541"
```

By default, `seaq` will fetch the entire thread. Use `-t` or `--tweet` to get a single tweet.

```sh
seaq fetch x "1883686162709295541" --tweet
```

> Note:
>
> `seaq fetch x` requires setting the `X_AUTH_TOKEN` and `X_CSRF_TOKEN` environment variables. You can get these tokens by inspecting the cookies after logging in to your X account.

#### `--no-cache` and `--json`

All fetch commands support caching extracted results. Cached entries are stored in `cache.db` in your config directory and are valid for 24 hours. To disable caching, use the `--no-cache` flag.

Fetch commands also support outputting JSON. Use the `--json` flag to enable this.

Please see `seaq fetch --help` for more information.

### Ollama support

`seaq` supports using local Ollama models through the `ollama` provider prefix.

```sh
# Use a local Ollama model
seaq -m ollama/llama2 input.txt
```

### Connection

`seaq connection` allows you to manage OpenAI-compatible API endpoints. This is useful when you want to use alternative providers that implement the OpenAI API specification.

```sh
# Create a new connection
seaq connection create groq --url https://api.groq.com/openai/v1
```

This will add a new connection to the config file.

```yaml
connections:
    - base_url: https://api.groq.com/openai/v1
      provider: groq
model:
    name: anthropic/claude-3-5-sonnet-latest
pattern:
    name: take_note
    remote: https://github.com/danielmiessler/fabric
    repo: /home/nt54hamnghi/.config/seaq/patterns
```

```sh
# List all configured connections
seaq connection list

PROVIDER      BASE URL                          ENV KEY
groq          https://api.groq.com/openai/v1    GROQ_API_KEY
openrouter    https://openrouter.ai/api/v1      OPENROUTER_API_KEY
```

The `ENV KEY` column tells you the environment variable that needs to be set for the connection to work.

```sh
# Remove a connection
seaq connection remove groq
```

Once a connection is created, you can list all models.

```sh
seaq models list
```

### Download remote patterns

To add new patterns from a remote repository, `seaq` expects the repository to have a top-level `patterns` directory with one or more patterns.

The directory structure should look like this:

```sh
patterns/
├── improve_prompt
│   └── system.md
├── prime_mind
│   └── system.md
└── write_blog
    └── system.md
```

By default, `seaq` will use `pattern.remote` in your config file as the remote repository. However, you can overwrite this with the `--remote` flag.

```sh
seaq pattern add improve_prompt --remote https://github.com/danielmiessler/fabric
```

## Acknowledgments

- Special thanks to Dr. Justin Sung for the inspiration
- [`fabric`](https://github.com/danielmiessler/fabric) project for inspiring seaq's design and functionality
