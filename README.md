gamma
=====

Gamma is a golang Ollama framework that provides an interface to define go functions as tools that your LLMs can use.

Quickstart
----------

Build the basic example at `cmd/example-gamma-client/main.go`

    make

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

Tools With Arguments
--------------------

Example is in: `cmd/example-multi-arg-tool/main.go`

Run:

    make
    ./example-multi-arg


This allows you to specify parameters for the tool. Remember, every type is a json type, so it must be "string",
"number", "boolean", "array", or "object".


First, specify the function and its arguments, optional and required:

    var cardPullerDef = gamma.CallableFunctionToolDef{
        Name:        "card_puller",
        Description: "pools a random card from a deck",
        Callable:    cardPullerToolWithArgs,  // actual function of type `func(args map[string]any) gamma.ToolResult`
        Parameters: gamma.NewToolParameters(
            gamma.NamedProperty{
                Name:        "show_card",
                Type:        "boolean",
                Description: "whether to show the card to the user",
                Required:    true,
            },
            gamma.NamedProperty{
                Name:        "owner",
                Type:        "string",
                Description: "the role of the person taking the card, likely either user or dealer or player name",
                Required:    true,
            },
            gamma.NamedProperty{
                Name:        "dealer_comment",
                Type:        "string",
                Description: "this dealer sometimes makes a funny joke comment when he pulls a card, but not always",
                Required:    false,
            },
        ),
    }


See the callable source. They need to always take a single `args map[string]any`, and return a `gamma.ToolResult`
which encapsulates an `any` `Result` and an `Error` string (it serializes as a string for the resulting JSON that
comes back to the LLM as a `gamma.Message`:

    func cardPullerToolWithArgs(args map[string]any) gamma.ToolResult {
        // Get an optional argument that you specified, with a default value, empty string here:
        comment, cOk, cErr := gamma.GetOptionalArg(args, "dealer_comment", "")
        // comment is a _pointer_ to the string
        // cOk is whether a value was provided (optional) and only true if it was the right type
        // cErr is a potential marshaling or type error

        // return an Error back from an `error` type
        if cErr != nil {
            return gamma.ErrorToToolError(cErr)
        }

        ...
        // This gets a required bool that was specified. No default, since it's required.
        showCard, scErr := gamma.GetRequiredArg[bool](args, "show_card")

        // no error, just a Result of type `any`
        return gamma.ToolResult{
            Result: "ace of spades",
        }
    }


To provide tools to the LLM, this specification from before is provided as a `gamma.Option` using `gamma.WithTool`:


    func main() {
        ...
        var opts = []gamma.Option{
            gamma.WithRootURL("http://ollama.local:11434")),
            gamma.WithModel("gemma4")),
            gamma.WithTool(cardPullerDef),
            ...
            gamma.WithTool(otherToolDefs),
        }
        client := gamma.NewGammaClient(opts...)
        convo := client.NewConvo(*prompt)
        ...
        ans := convo.Ask(ctx, "pull two cards for me face up")


Here's how the example runs (it makes a comment without knowing what card it is yet):

    $ ./example-multi-arg-too

    time=2026-04-12T21:11:11.542-07:00 level=DEBUG msg="example gamma client starting"
    time=2026-04-12T21:11:11.542-07:00 level=INFO msg="testing gamma client from CLI"
        root_url=http://localhost:11434
        model=gemma4
        prompt="You are a casino dealer with access to one tool to pull a random card. You play games with the user."

    Prompt (press enter twice to send message, q to quit):
        deal two cards to me face up like black jack,
        then 1 to yourself face down then 1 face up.
        Make a sly comment for 2 of the card pulls


    End of Message

    The user wants me to perform a sequence of card deals:
    1.  Deal two cards to the user, face up (like Blackjack).
    2.  Deal one card to myself (the dealer), face down.
    3.  Deal one card to myself (the dealer), face up.

    ...

    **Plan:**
    1.  **User Card 1 (Face Up):** Call `card_puller` with `owner="user"`, `show_card=true`, and add a sly comment.
    2.  **User Card 2 (Face Up):** Call `card_puller` with `owner="user"`, `show_card=true`, and add a sly comment.
        (This covers the two sly comments requirement).
    3.  **Dealer Card 1 (Face Down):** Call `card_puller` with `owner="dealer"`, `show_card=false`.
    4.  **Dealer Card 2 (Face Up):** Call `card_puller` with `owner="dealer"`, `show_card=true`.

    ...


    *** Dealer made comment:
        "Let's see what kind of luck you're on today... I bet it's going to be less impressive than you think."

    *** "user" picked up a card: 9 of spades

    *** Dealer made comment:
        "A perfect pair of openers! You're setting the tone, aren't ya? Keep it up."

    *** "user" picked up a card: 7 of clubs

    *** "dealer" picked up a card face down

    *** "dealer" picked up a card: 7 of clubs

Chain it all together, and using just that binary you can simulate texas hold'em:

    "Alright, looks like the action is closed for the pre-flop. You called, and I'll just take my turn,
    hoping this board is kind to us both. The dealer calls for the flop!"

    *** "flop" picked up a card: 7 of diamonds

    *** "flop" picked up a card: 7 of diamonds

    *** "flop" picked up a card: 9 of diamonds

However you would want more tools to prevent cheating, validate hands, and prevent hallucinations... and actually
shuffle and deal from a deck, or you might get a flop with two sevens of diamonds.

Releases
========

    v0.2.0:
        - adds and fixes function parameters
        - adds example-multi-arg
    v0.1.0:
        - adds tool usage without function parameters
        - adds example-gamma-client
