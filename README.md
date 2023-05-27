# Go Get Er Down

## Purpose
This tool can be used to test your website's resilience against DOS attacks.  
Once launched, the tool will ask for the following information:
- Website to target
- Number of concurrent connections to use during the test attack
- Total number of requests to send

What makes this tool a little different is that it will first navigate the given website for a few minutes in order to gather as many unique internal links as possible. This gives it the advantage to have many (potentially tens of thousands) unique URLs to hit on the domain once the attack begins at high concurrency levels.

## How to use it
```cmd
go build
./go-bring-er-down
```

## Output
```
What website do you want to bring down? https://test.mywebsite.ca
How many concurrent requests? 350
How many total requests would you like to send? 1000000

Alright... Let's get https://test.mywebsite.ca down. We'll do 350 concurrent HTTP calls for a total of 1000000 requests.

Warming up...
Casually navigating website to capture as many distinct URLs as possible... (this can take a few minutes)

We've got 14895 distinct URLs to work with, let's get ready...

GO!

DONE
```

## Disclaimer
This tool is intended for educational purposes only.  
It should be used to test your own website.