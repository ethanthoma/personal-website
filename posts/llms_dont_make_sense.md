# LLMs don't make sense

> tldr; AGI won't come from just bigger text-based LLMs. Progress demands a
> shift towards models grounded in real-world interaction, guided by dynamic
> values, and built on more versatile, data-aware architectures.

Picture two computers, Timmy and Billy.

They communicate by sending signals, bit by bit.

Now, imagine that you can't access these devices directly. In fact, you don't
even know precisely what they do, or if they're even computers as we understand
them. All you can observe is the network and the traffic flowing between them: a
few bits from Timmy, a few bits from Billy.

But perhaps you want to talk to Timmy or Billy yourself. You want to join their
conversation. You have all of their communications. So, what's your strategy?

Well, let's say we train a neural network. Since we don't know exactly what
Timmy and Billy are, we try to approximate their behavior. Off you go to train
it.

And it works, to a degree. You can "talk" to them, and they reply to you. Your
dataset grows. You train bigger models. You throw more compute at it. Rinse and
repeat.

Flush with more data and more compute, you keep laying your bricks, building a
grander tower. Each neural layer a new layer of brick, reaching higher and
higher. A veritable Tower of Babel. No matter how high our computational tower
grows, it will never allow us to truly grasp the internal essence of Timmy or
Billy. we're just getting a taller vantage point with the same fundamentally
limited view.

## Language Modeling is a Sub-task

> We were all of us cogs in a great machine which sometimes rolled forward,
> nobody knew where, sometimes backwards, nobody knew why.
>
> — Ernst Toller

The language modeling part of a Large Language Model is a sub-task.

A necassary condition for AGI, sure, but hardly sufficient.

I think, we lost sight of this.

We have clear evidence that exceptional language modeling doesn't require AGI.
Current models like Claude and ChatGPT generate text with such fluency that
their abilities can *feel* tantalizingly close to true understanding—an illusion
of approaching AGI asymptotically. Indeed, research on scaling large
conversational agents (like the work on models such as Meena) often highlights
how more scale leads to better performance on language tasks. And it's true:
continued scaling will undoubtedly yield even more refined language models.
However, this path doesn't lead to AGI. It leads to an ever-increasing compute
bill for what is, at its core, an exceptionally sophisticated token guesser.

We have clear evidence that language modeling doesn't need AGI. Models like
Claude and ChatGPT and so on model language exceptionally well. In fact, it
*feels* asymptotically close to AGI. If anything, the
[ChatGPT paper](https://arxiv.org/pdf/2001.08361) almost implies this. And it's
true, continued scaling will undoubtedly yield even more refined language
models. But it won't get to AGI. It leads to an ever-increasing compute bill for
what is, at its core, an exceptionally sophisticated token guesser.

## The Human Experience

> "You never really understand a person until you consider things from his point
> of view, until you climb inside of his skin and walk around in it."
>
> — Harper Lee

Text is a terrible communication protocol.

So much is inherently missed; so much vital information is simply lost. Consider
all that's conveyed through body language, facial expressions, tone of voice,
environmental context, and shared physical presence. None of these neatly fit
into a string of characters.

More so, effective communication is challenging even for us humans, despite
possessing the full spectrum of human experience and being literally built for
these social interactions. How do you expect a model to adequately grasp or
replicate that experience solely from text?

I see two core issues. First, the profound lack of embodied human experience for
our models. For a person to truly understand another's experience, they often
need to engage in similar activities, perceive similar environments, and
essentially try to align their sensory inputs. Reading the same texts is a
start, but it pales in comparison to doing the same things, saying the same
things, and striving for a **similar sensory experience**.

This is where multi-modality becomes crucial, yet it too is insufficient on its
own. Models need to adapt, interact, and explore their environments. We need
models that can, metaphorically or even literally, try to put small Legos in
their mouths—to learn from unscripted, curiosity-driven engagement with the
world.

**Humans perceive first, then communicate!**

The second core issue is implicit knowledge. As humans grow, we interact, we are
taught, and we absorb a complex web of values, social norms, and cultural
commonsense. This deeply ingrained understanding shapes what we know, what we
infer, and how we communicate. They are apart of why text is a terrible medium.
Admittedly, there have been large efforts to combat this, especially across
modalities.

In my view, tackling these issues requires a multi-faceted approach. Firstly, we
need methods for models to acquire socially implied knowledge more dynamically,
perhaps through mechanisms like test-time learning triggered by **surprise
mechanism** (i.e. when model expectations deviate from observations). Secondly,
we need to develop robust value systems or stores, allowing models to represent
and weigh what they deem important. Which naturally leads me to...

## Values and Reinforcement Learning

> "I used to rule the world..."
>
> — Chris Martin

Reinforcement learning (RL) was king for much of 2000s. Before AlexNet's big run
and the eventual transformer hegemony, RL was the coolest kid on the block. And
it will outlive deep conv-nets and transformers.

**Value systems are an integral part of decision making**. Economics, at least
partially, captures this aspect of our nature; it's why I consider economics and
game theory vital for machine learning research. For instance, you might prefer
fruit A to fruit B (or, like me, you might not be a big fruit eater at all).
Yet, after consuming a certain amount of fruit A, your preference will likely
shift towards fruit B (marginal utility goes brrr).

This is important. In fact, I think it's really important. Even in language,
there's an inherent entropy. You know that too. You probably can guess how a
close friend might reply to your text under normal circumstances. But what if
they're stressed? Or what if they've been bombarded all day by similar requests?
Their response, and the underlying _value_ of engaging with your text, would
almost certainly change.

The decisions we make are heavily influenced by our values. And they change.
They change by whims and emotions and our environment. They impact what we say
and do. Most importantly, its variance will give a base entropy to our language
modeling.

Reinforcement learning offers a framework to model this. In part, it already
has. Yes, RL is reward-based, but aren't we? Food, exercise, social acceptance
all trigger our ingrained reward networks.

We need models that can interact with their environment, continuously updating
their knowledge and internal values. RL will hand it to us on a silver platter.

There are beautiful papers out there that explore this, like the
[Worldformer](https://arxiv.org/pdf/2106.09608) model which demonstrate that
models can update their knowledge and operate over it 'live'. It's a path I hope
many more researchers will explore.

## Architecture is King

> "Go to, let us build us a city and a tower, whose top may reach unto heaven;
> and let us make us a name, lest we be scattered abroad upon the face of the
> whole earth."
>
> — The Book of Genesis

Transformers have one flaw. They're not the perfect model (Okay, maybe there may
be more than one flaw deep down). I feel like people forget what transformers
were made for. Have we really reached a point where we take a rich 3D scene,
flatten it to 2D images, and then...sequence that into 1D sequence of "patches",
all to feed a transformer? _Huh!?_ Are we so fixated on benchmark scores that
our model architectures bear little resemblance to the inherent structure of our
data?

It feels like only recently, as our 'mega-transformers of doom' begin to collide
with the hard realities of the fiscal-compute ceiling, that a collective
realization is dawning: maybe, just maybe, we can make better models. Models
that actually cater to the task at hand. In the pre-transformer era we used to
build assumptions **into** our models. In fact, a lot spaces still do! To grand
effect. It's time to revisit that philosophy, to consciously 'bake' our
understanding of the data and the underlying processes back into our model
architectures. Attention mechanisms are powerful, but they are **not** all you
need.

The latent dimensionality of our models is huge. So large in fact, imposing
constraints would probably hurt the model performance no more than a random seed
does. We can design latent spaces that respect the inherent structure of our
data. We can build in robust assumptions. Doing so can lead to models that train
faster and generalize better because they conform more closely to the problem's
nature. People already know this too. We already constrain our latent spaces.
But there's so much more we can do, if we're willing to venture beyond current
orthodoxies and truly explore.
