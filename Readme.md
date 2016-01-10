# Cornerbot

## Setup

Rename `bot.db.skeleton` to `bot.db` as well as `conf.json.skeleton` to
`conf.json`. Then run `go build`, and `./cornerbot`.

## Customization

If you want custom action commands or message commands, add them to the
database in the commands table. Currently supported types of custom commands
are 'action' and 'message'.

If you wanted the bot to respond to
'!poke mikey' with 'cornerbot pokes mikey with a stick!', you would execute the
SQL statement `insert into commands(name, message, type) values("poke", "pokes
%s with a stick", "action")`.

If you wanted '!yell' to have the response 'cornerbot yells AHHHHHH', you'd execute
`insert into commands(name, message, type) values("yell", "yells AHHHHHH", "action")`.

If you wanted '!motd' to have the response 'cornerbot: Do not forget there is a full
moon tonight!', you'd execute the statement 
`insert into commands(name, message, type) values("motd", "Do not forget there is
a full moon tonight!", "message")`.

If you wanted 'robert: !love' to result in 'cornerbot: I love you too, robert!',
you'd execute
`insert into commands(name, message, type) values("love", "I love you too, %s!", "message")`.
