## What is this?

Have you ever needed to extract the emails of all the people you've
discussed certain topics with?

Maybe to bootstrap a mailing list, a newsletter, or to send them an update?

**`gmailcrawl` is a command line tool, written in go, to do extract all email
addresses involved in threads you select.**

It only works with GMAIL accounts, since it uses GMAIL APIs.

Unlike many online services that allow you to scan your emails and extract
addresses, it does not require you to authorize third parties to access your
mailbox.

Your security token is stored on your machine, and only shared with
the very simple `gmailcrawl` code you run yourself.

Once you install it, you can, for example, run:

    ./gmailcrawl --query="robots championship" --limit=1000

to get gmailcrawl to scan all of your emails that talk about "robots
championship", and print to standard output a list of unique email addresses.

You can also use the `--whitelist` or `--blacklist` flag to discard some
addresses based on regular expressions.

The default `--blacklist` automatically excludes some known email addresses
you probably don't care about (for example, twitter notifications or
updates to your google docs).

## Installation

1) Install the `go` build and runtime environment. On a `Debian GNU/Linux`
   system, run:

       apt-get install golang

2) Compile and install gmailcrawl:
    
       go get github.com/ccontavalli/gmailcrawl


3) Profit:

       $GOPATH/bin/gmailcrawl --query="robots championship" --limit=1000
