# LLMs Suck

Picture two computers.

They can send little signals to each other over the so called "network".

We know computers have more information than what is communicated over the
network. Could be a lot of internal information, a lot of processing, etc.

Now imagine you are trying to train a neural network. The goal? To emulate a 
computer. The data? Whatever is sent over the network...

You may be an optimist. Fair. I am not though. I simply cannot see how the 
network will become a general purpose computer from looking at network data. And 
for that reason, I don't see how LLMs will learn to be AGI by looking at text.

## The Human Experience

> "How does an LLM walk in my shoes for day?"

Language modeling is a subtask. However, I feel like we lost site of this.

AGI *would* be able to model language, but language model doesn't need AGI.

We know this to be true. We have Claude and ChatGPT and LLama and so on. They 
model language exceptionally well. It feels almost asymptotically close to AGI,
but it won't get there. Scaling a transformer model to trillions of params does 
not give you sentience, it gives you a big compute bill for a great token guesser.

Don't get me wrong, language modeling is useful. Really useful. That's why the 
industry is throwing billions at it. But it is not the final boss. You know what 
the real language "model" is?

Literally any person.

That's because the actual goal is the human. You want to model the language 
maker, not predict the language maker's next move. These are **not the same**.

Part of this comes from text being a terrible communication protocol. So much is
missed, so much is conveyed through other mediums like body language and facial 
expressions and tone and so on. Theres a reason people still struggle with 
communication in their adulthood. It's hard. It's hard for humans who have the 
human experience. To propose that a model can model human experience from text 
is insane to me. 

Humans already have a human experience, and even with other mediums such as voice,
I don't think humans are all that capable of modeling the experience of another.
And we come with the best pre-training for it too!

The best for a human to experience the experience of another is to literally 
engage in the same experiences. Yes, read the same things, but also do the same 
things, say the same things, match your senses to theirs. Fundamentally, you 
need to have a similar sensory experience. 

**Humans percieve first then communicate!**

## Values and RL

> "Reinforcement learning: I used to rule the world..."

Reinforcement learning (RL) was king for much of 2000s. Before AlexNet's big run 
and the eventual transformer hegemony, RL was the coolest kid of the block. And 
I think, it will outlive deep conv nets and transformers.

Value systems are an integral part of decision making. Economics captures this 
part of us, at least partially (and why I think it and game theory are important 
for machine learning research). You may like fruit A more than fruit B (or 
you're like me who doesn't really eat fruit) but after giving you _x_ fruit A, 
you will probably want fruit B more...

And this is important. Really important. OpenAI knows that there is a base 
entropy in language modeling. You know that too. Think about it. You probably 
can guess what someone you're close to would say if you text them. But would 
that change if they were more stressed? What if they normally like your texts 
but today they got hammered by a bunch of people asking them the same thing? 

I find the smartest someone is, philosophically not academically, the more 
self-consistent their value system tends to be. The decisions we make are 
heavily influenced by multiple factors like emotion our values. How your values 
changed affect how you interact with the environment, including communication, 
which can not be predicted on average, like a LLM tries to do.

I don't think we want an average-predicting model, we want AGI, something I hope 
is a bit better than the average person.

## Architecture is King

> "Go to, let us build us a city and a tower, whose top may reach unto heaven; 
> and let us make us a name, lest we be scattered abroad upon the face of the 
> whole earth." -The Book of Genesis

OpenAI has very succesfully [psyoped](https://en.wikipedia.org/wiki/Psychological_operations_(United_States)) 
us into thinking scaling was king. More parameters, more compute, more data, more
more more. But, I think they're wrong. Transformers can scale really well. And 
so we do scale them. But if you think transformers are the final architecture 
then go off and spend your trillions on training your single model.

Transformers have one flaw. They're not the perfect model. Okay, maybe that's 
more than one flaw deep down but I feel like people forget what we made it for.
We are at the point where we take a 3D scene, project it onto 2D (pictures FTW) 
and then...sequence it into 1D just so we can use a transformer? Huh!? We are so 
lost in desire for better benchmark performance that our models have almost 
nothing to do with our data? Do we just hate statistics that much?

I feel like on recently, as our mega-transformers of doom hit the fiscal-compute 
ceiling have we come to terms with maybe, just maybe, we can make better models.
In the pre-transformer era we used to build assumptions **into** our models. We 
have the physics formulas for light and projections and so on, so we added this 
into our models and we got better models, or at least quicker to train models.

I think this underlines why AlexNet era and transformer is so different. A CNN 
is blazingly fast to train and takes way less data but sure, transformers do 
predict a bit better. I think we need to go back to where we construct our 
knowledge of the data or the process back into the model. Attention mechanisms 
are great, they are just **not** all you need.
