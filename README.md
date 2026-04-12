gamma
=====

Gamma is a golang Ollama framework that provides an interface to define go functions as tools that your LLMs can use.

Quickstart
----------

Build the basic example at `cmd/example-gamma-client/main.go`

    make example

Run it:

    ./example-gamma-client

The most basic example just provides one very simple tool called `card_puller` with no arguments, and specifies that
the LLM assistant is a casino dealer. You can ask it to pull cards, or even try to play a basic game of no hold'em
but you would have to be very specific with it in the questions you pass to it, like so:

    I want to play texas no limit hold'em. You need to run the card pulling tool twice for my cards, twice for
    yourself, then 3 times for the flop, 1 time for the turn, and 1 time for the river.

    You will then tell me what my pocket cards are, and action is on me. I will either call with 1 chip to match
    big blind's 2 chips, or fold or bet. Then you will tell me whether you call/raise/fold/etc, then tell me the
    flop if we checked.

You press enter twice to send a message, and the message can be just `q` if you want to quit. It will say
"End of Message" if it detected your two newlines.

A better poker player would have more tools to track state and accept bets and handle the math in pure logic. You will
notice that the LLM is likely to hallucinate, or end the game too early, or pretend it's at the final river card when
the game hasn't progressed there yet.

But as an example, that source shows you how to initialize a client, provide a tool, and interact with the streaming
service.
