# Signal LLM Chat server

Go module that connects to a Signal CLI API Container and relays messages to and from a local LLM.

## Basic Architecture
This is a very simple program, it simply consumes the output for a single number from [Signal REST API](https://github.com/bbernhard/signal-cli-rest-api) and relays it to [Open WebUI](https://github.com/open-webui/open-webui). Ensure the Signal REST API is running before running this module.

## Requirements
### [Signal REST API](https://github.com/bbernhard/signal-cli-rest-api)
Run this in json-rpc mode with port 3001 forwarded.

### [Open WebUI](https://github.com/open-webui/open-webui)

I use Open WebUI because it offers a REST API (though it turns out to be poorly documented). In the future, I will implement the `completed` API call to persist chats to the UI. Run this with port 3000 forwarded.

## Build and run
```shell
# Run these commands form the /src folder
$ go build .
$ cp ../.env ./
# Edit .env file accordingly
$ ./signal-llm-chat
```

## Usage
Once running, simply send a message to the number you setup on Signal to initiate a new chat. This will establish a chat ID that is mapped to your number and 

### Text commands
There is a limited set of commands supported though leading bangs

**Models**  
!m - Return model currently being used for your chat  
!m list - list all available models  
!m load [model-name] - Change the model being used

**Web Search**  
!w - toggle web search (default off unless specified in .env)  
!w [true | on | 1] - enable web search  
!w [false | off | 0] - disable web search

## Configuration
Configuration is achieved through a typical .env file, an example of which is in the top level of this repository.

.env options:
``` bash
OPENWEBUI_API_KEY=// Open WebUI Page: https://docs.openwebui.com/getting-started/api-endpoints/

OPENWEBUI_CHAT_ID=// Get this from the URL in Open WebUI of the chat you want to use

OPENWEBUI_MODEL=// Model name and size, i.e. mistral:7b

OPENWEBUI_MODEL_DEFAULT=// Default model for new chats

OPENWEBUI_URL=// In the form of [host]:[port] without the protocol, i.e. localhost:3000, 192.168.1.12:3000

SIGNAL_NUMBER=// Must include '+[country code][7-digit number]'. Ex: +13549687

SIGNAL_URL=// In the form of [host]:[port] without the protocol, i.e. localhost:3000, 192.168.1.12:3000

DEBUG=// Set to 1 for extra logging. Note: This will print anything in the text message, so be aware of any sensitive content while this is enabled.
```

## Ongoing Features
These are features that have no definition of done, but will likely be further developed as I think of things
- [x] Server controls via text (Changing models, updating prompt, etc.)

## Future Features (In no particular order)
- [x] Implement `.env` file
- [ ] Figure out Open WebUI `completed` API call to persist chats in UI
- [X] Implement sender look up with chat ID
- [X] Create new chat for senders no in lookup table
- [ ] Implement Signal group chats (Will only respond to @bot-name)
- [x] Enable web search
- [x] Enable file attachments to support RAG
- [ ] Implement text formatting

