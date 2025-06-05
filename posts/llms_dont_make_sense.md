# LLMs don't make sense

Picture two computers, Timmy and Billy.

These two can communicate. They send little signals, bit by bit.

Now, imagine that you can't access these computers. In fact, you don't even know
what they can do. They might not even be computers! All you can see, is the network 
and its traffic between them.

A couple bits from Timmy, a couple bits from Billy.

But, maybe you also want to talk to Timmy or Billy. You also want to join 
communicate in this network. You have all their communications. So what do you do?

Well, lets train a neural network. We don't know exactly what Timmy and Billy are,
so we can approximate them. Off you go to train it.

It works. A little bit. You can "talk" to them, and they reply. Your data 
grows. You train bigger models. You use more compute.

Now, you may be an optimist. Fair. You might think if we can just get more data, 
more compute, maybe then we can emulate a Timmy or a Billy. But can you? Do you really 
think, you will emulate a computer? Just from network data? I don't.

Computers are complex beasts. We made them efficient and optimized to perform 
the things it does. Sure, a trillion parameter model might be able to do 
computer-like things. But would it be a computer? Would more parameters make it
a computer? How many orders of magnitude more compute would you need to make it 90% 
as good?

We can approximate it, sure. But it won't be better. Hell, by the time it trains, 
it will probably be outdated already and for that reason, I don't see how LLMs 
will become AGI, especially if we just learn from text.

## Language Modeling is a Sub-task

The language modeling part of a Large Language Model is a sub-task.

A necassary condition for AGI, sure, but hardly sufficient.

I think, we lost sight of this.

We know language modeling doesn't need AGI. We have Claude and ChatGPT and so on. 
They model language exceptionally well. In fact, it feels almost asymptotically 
close to AGI. If anything, the [ChatGPT paper](https://arxiv.org/pdf/2001.08361) 
almost implies this. If we keep scaling, we will get better at language modeling.
But it won't get to AGI. It gives you a big compute bill for a great token guesser.

Language modeling is useful. That's why the industry is throwing billions at it. 
But it is not the final boss. You know what the real language "model" is?

You. Me. Literally anyone.

The real goal of language modeling is to model the language maker, a human, not 
just predict the language maker's next move. These are **not the same**.

## The Human Experience

> "You never really understand a person until you consider things from his point of 
> view, until you climb inside of his skin and walk around in it."
>
> — Harper Lee

One of the main reasons language modeling is not the gateway drug to AGI, is because
text is a terrible communication protocol. So much is missed. So much information 
is simply lost. So much is conveyed through other mediums like body language and facial 
expressions and tone and so on. 

Communication is hard. Even for humans and humans have the human experience! 
They're literally built for human-interaction. To propose that a model can model 
human experience from text is frankly, insane to me. 

I think there are two issues at hand. One, the lack of human experience for 
model. The best way for a human to learn another's human experience is to literally 
engage in the same experiences. Yes, that includes reading the same things, but 
also means to do the same things, say the same things, match your senses to theirs. 
Fundamentally, you need to have a **similar sensory experience**.

This is why multi-modality is important. But also, not sufficient. Models need to 
adapt and interact. We need models that can put small Lego's into their mouth.

**Humans perceive first then communicate!**

The second issue is implicit knowledge. Humans, as we grow and age, engage with 
others and are taught by others. We develop values, norms, and cultural commonsense.
These affect what we know and what we say. They are apart of why text is a terrible 
medium. Admittedly, there have been large efforts to combat this, especially across
modalities.

In my mind, the solution will be in two parts. First, adding socially implied knowledge 
to the model, probably through a test-time learning with a surprise mechanism (i.e. 
when model expectations deviate from observations). Secondly, a value store/system, that 
allows models to represent what they value and by how much. Which leads me to...

## Values and Reinforcement Learning

> "I used to rule the world..."
>
> — Chris Martin

Reinforcement learning (RL) was king for much of 2000s. Before AlexNet's big run 
and the eventual transformer hegemony, RL was the coolest kid on the block and I 
think, it will outlive deep conv-nets and transformers.

**Value systems are an integral part of decision making**. Economics captures this 
part of us, at least partially; it's why I think economics and game theory are 
important for machine learning research. You may like fruit A more than fruit B (or 
you're like me who doesn't really like fruit) but after giving you _x_ fruit A, 
you will probably want fruit B more...

This is important, in my opinion. In fact, I think it's really important. OpenAI 
knows that there is a base entropy in language modeling. You know that too. 
You could probably can guess what someone you're close to would say if you text 
them. But would that change if they were more stressed? What if they normally like 
your texts but today they got hammered by a bunch of people asking them the same thing? 

The decisions we make are heavily influenced by our values. And they change. They
change by whims and emotions and our environment. They impact what we say and do.
Most importantly, its variance will give a base entropy to our language modeling.

RL can do this for us. RL has done this for us, in part. Sure, it's reward based, 
but aren't we? Food makes us happy, exercise too, as well as being socially accepted. 
A lot of our reward networks are biologically ingrained and in the context of RL,
will likely require us to specify very targeted reward systems.

We need to let our models interact with the environment, and update their knowledge 
and values. RL will hand us that on a silver platter. There are beautiful papers out 
there that explore it too. Some papers like the [Worldformer](https://arxiv.org/pdf/2106.09608) 
show us that we can update our knowledge and operate over it "live". I hope many 
more researchers explore this path.

## Architecture is King

> "Go to, let us build us a city and a tower, whose top may reach unto heaven; 
> and let us make us a name, lest we be scattered abroad upon the face of the 
> whole earth." 
>
> — The Book of Genesis

OpenAI has very successfully psyoped us into thinking scaling was king. More 
parameters, more compute, more data, more more more. But, I think they're wrong. 
Transformers can scale really well. So we do scale them.

Transformers have one flaw. They're not the perfect model. Okay, maybe that's 
more than one flaw deep down but I feel like people forget what we made it for.
We are at the point where we take a 3D scene, project it onto 2D (pictures) 
and then...sequence it into 1D just so we can use a transformer? Huh!? We are so 
lost in desire for better benchmark performance that our models have almost 
nothing to do with our data?

I feel like only recently, as our mega-transformers of doom hit the fiscal-compute 
ceiling have we come to terms with maybe, just maybe, we can make better models.
Models that actually cater to the task at hand. In the pre-transformer era we used 
to build assumptions **into** our models. In fact, a lot spaces still do! To grand
effect. I think we need to go back to where we bake in our knowledge of the data or 
of the process back into the model. Attention mechanisms are great, they are just 
**not** all you need.

The dimensionality of our models are huge. So large in fact, imposing constraints 
probably hurt the model performance no more than a random seed does. We can constrain 
our latent spaces to our data. We can build in assumptions. We can make our models
train faster and conform better. People already know this too. We already constrain 
our latent spaces. But there's so much more we can do, if we are willing to explore 
more.
