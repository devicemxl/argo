# Cross-Package CGO Compatibility

One of the trickiest aspects of working with Argo was ensuring its compatibility across different Go packages. CGO generates package-specific types, such as *argo._Ctype_char and *main._Ctype_char, which are not directly compatible with each other. 

To address this challenge, I implemented a helper function, This helper function allows seamless interaction between different packages. Here’s how you can use it in your application

[image1](https://miro.medium.com/v2/resize:fit:4800/format:webp/1*L84hlP-2s-SolriniGqLrg.png)

[image2](https://miro.medium.com/v2/resize:fit:4800/format:webp/1*yE0OFNEDq341JaVYUul4mw.png)
