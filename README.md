# Trading 212 Parser

This parser should parse the buy/sell transactions from history files exported from Trading 212 to output the yearly Irish CGT liability (best effort/close enough basis) using the FIFO method to calculate profits.

**NOTE:** I take no responsibilty if this does not calculate your taxes correctly. Always consult a tax advisor for legal matters. 

## Running

* Populate the `./configs/config.json` file with a map of `Year` to `Path` of the history file exported from Trading 212
* Run `go run cmd/main.go -config configs/config.json`
* Run `go run cmd/main.go --help` for usage

The output text will show you your estimated tax liability for all the years.

It follows
- FIFO as a default mechanism
- LIFO when a stock is sold after being bouth within the last 4 weeks (and taking FIFO when applicable in this case)
- Currency exchange fees are proportionally taken when needed (partial shares being sold)
- Currency exchange losses are reflected in the transaction history itself by the vertue of everything being converted to Euros
    - This is to say that no specific provisions are made to handle these cases
- **READ THE NOTES FOR EXCEPTIONS**

## Explanation

The below is from a Revenue MyEnquiries correspondance

> If the assets being sold are Shares/Stocks/Cryptocurrencies
> Where the shares being disposed of are of the same class (e.g. ordinary shares) the general rule is that the first shares acquired are deemed to be the first sold (i.e. FIFO method - first in - first out).
> 
> Unless:
> 1. Acquisition of shares within 4 weeks of disposal
> In this case, if a loss occurs on the initial disposal, then this loss can only be offset against a gain on the sale of shares of the same class which were purchased within 4 weeks of that sale.
> 2. Disposal of shares within 4 weeks of acquisition
> The FIFO rules do not apply to this 'bed and breakfast' transaction. Instead, the last in/first out ("LIFO") rule will apply so that the shares acquired within 4 weeks prior to the sale of shares of the same class are deemed to be sold first.
> 
> Where an individual incurs a loss on disposal of shares, that loss can be deducted from chargeable gains in the same period. If the loss is greater than the chargeable gains, the remaining loss can be carried forward against future chargeable gains.

This, along with information from [this self-initiated reddit thread](https://www.reddit.com/r/irishpersonalfinance/comments/1gamjte/shares_cgt_calculations_with_their_caveats/) and [associated tax information](https://www.irishstatutebook.ie/eli/1997/act/39/section/581/enacted/en/html#sec581), I conclude that these rules apply (and I will try to implement these best I can with no guarantees what so ever)

For the calculation of gains:
* FIFO applies by default
* LIFO applies if the same stock is being sold within 4 weeks of aquisition
* Losses from sales when the same stock is bought again within 4 weeks:
    * If you sell `100` shares and rebuy all of them
        * Loss from the sale can only be offset against gains from the 100 shares bought and nothing else ("wash sale", this is not a general allowable loss because of it)
    * If you sell `100` shares and rebuy `40` of them within 4 weeks (no more buys after that)
        * Loss from `100 - 40 = 60` shares can be offset as general allowable loss against any other CGT
        * Loss from the `40` shares rebought can only be offset against the sale of these `40` shares

ChatGPT was used to help understand the above as well, for context:

<details>
<summary>ChatGPT conversation</summary>

***You said:***
Explain this in different words to me 

***ChatGPT said:***
Of course! Could you share the text you’d like me to rephrase?

***You said:***
 For the purposes of the Capital Gains Tax Acts, where the same person in the same capacity disposes of shares of the same class as shares which such person acquired within 4 weeks preceding the disposal, the shares disposed of shall be identified with the shares so acquired within those 4 weeks.

***ChatGPT said:***
Under the Capital Gains Tax Acts, if a person sells shares of the same class that they bought within the previous four weeks, those sold shares will be matched with the shares bought during that four-week period.

***You said:***
Give me an example

***ChatGPT said:***
Sure! Here’s an example to clarify this rule:

Let's say:

Sarah owns 100 shares of XYZ Corp.
On March 1st, she buys an additional 50 shares of XYZ Corp.
On March 20th, she sells 50 shares of XYZ Corp.
Under the Capital Gains Tax Acts rule, the 50 shares she sells on March 20th will be considered the same shares she bought on March 1st, since this purchase happened within four weeks of the sale. This means any capital gains or losses Sarah has will be calculated based on the price she paid for the 50 shares on March 1st, not her original 100 shares.

This helps avoid "tax-loss harvesting" where investors could sell and quickly repurchase shares to create a tax benefit.

1/2

***You said:***
Does the mention of share classes mean any shares of the same class can be used in the above. Or does it need to be the same security?

***ChatGPT said:***
The mention of "share class" here means that the rule applies to shares of the same company and class. This means it isn’t enough that the shares belong to the same general class (like "common shares") across different companies; they must be the same security from the same company and of the same class within that company.

So, if you sell shares of XYZ Corp's common shares, the rule applies only to common shares of XYZ Corp that you bought within the previous four weeks. It would not apply to shares of a different company, nor to a different class of shares within XYZ Corp (e.g., if they had both Class A and Class B shares).

***You said:***
Can you explain this section now?

For the purposes of the Capital Gains Tax Acts, where the quantity of shares of the same class disposed of exceeds the quantity of shares of the same class acquired within the period of 4 weeks preceding the disposal, the excess shall be identified with shares of the same class acquired otherwise than within the period of 4 weeks.

***ChatGPT said:***
This section addresses situations where the number of shares sold is greater than the number of shares bought in the previous four weeks.

Here’s how it works:

If you sell more shares than you bought in the four weeks before the sale, the rule first matches the shares you bought within that four-week period. After those are matched, any remaining shares in the sale are then identified with shares that were bought outside that four-week window.

Example
Let's break it down with an example:

Sarah owns 200 shares of XYZ Corp.
On March 1st, she buys an additional 30 shares of XYZ Corp.
On March 20th, she sells 50 shares of XYZ Corp.
Here’s how the rule applies:

The first 30 shares of her sale on March 20th will be identified with the 30 shares she bought on March 1st, as they were purchased within the previous four weeks.
The remaining 20 shares from her sale (since she sold 50 shares total) will be matched with shares she acquired outside the four-week period—likely from her original 200 shares.
This rule ensures that recent purchases are considered first, with any remaining shares matched to older holdings.

***You said:***
can you explain this now

(3) Where a loss accrues to a person on the disposal of shares and such person reacquires shares of the same class within 4 weeks after the disposal, that loss shall not be allowable under section 538 or 546 otherwise than by deduction from a chargeable gain accruing to such person on the disposal of the shares reacquired; but, if the quantity of shares so reacquired is less than the quantity so disposed of, such proportion of the loss shall be allowable under section 538 or 546 as bears the same proportion to the loss on the disposal as the quantity not reacquired bears to the quantity disposed of.

(4) In the case of a man and his wife living with him—

(a) subsections (1) and (2) shall, with the necessary modifications, apply where shares are acquired by one of them and shares of the same class are disposed of within 4 weeks by the other, and

(b) subsection (3) shall, with the necessary modifications, apply also where a loss on the disposal accrues to one of them and the acquisition after the disposal is made by the other.

***ChatGPT said:***
Certainly! These sections deal with rules around capital losses when someone sells shares and then buys similar shares again shortly afterward. Here’s a breakdown of each part:

Section (3)
This section prevents a person from immediately claiming a tax loss if they sell shares at a loss and then buy back the same type of shares within four weeks. This is known as a "bed and breakfasting" rule, aimed at stopping people from creating tax losses to offset gains while still holding essentially the same investment.

Here’s how it works:

If you sell shares at a loss and buy back the same class of shares within four weeks, the loss from the sale is not allowed as a deduction under sections 538 or 546 for tax purposes.

Instead, that loss can only be used to offset a future gain when you eventually sell the reacquired shares.

If you reacquire fewer shares than you originally sold, only part of the loss will be restricted. Specifically, the allowed loss is proportional to the difference between the shares sold and those repurchased.
Example:
Sarah sells 100 shares of XYZ Corp at a loss.
Two weeks later, she buys 40 shares of XYZ Corp.
Under this rule:

Only a portion of Sarah's loss (corresponding to the 60 shares not repurchased) will be allowed as an immediate deduction.
The remaining loss (for the 40 shares reacquired) will be deferred and only allowed to offset any future gain from selling those 40 shares.
Section (4)
This section applies similar rules to married couples living together.

Subsection (4)(a)
Subsections (1) and (2) apply with necessary adjustments if one spouse buys shares and the other spouse sells shares of the same class within four weeks. So, the rules that match shares based on timing will apply across both spouses' transactions.
Subsection (4)(b)
Subsection (3)’s loss restriction also applies if one spouse incurs a loss on selling shares and the other spouse buys similar shares within four weeks. This means that the loss cannot be immediately claimed but is deferred, just as if the same person had done both transactions.
Example for Section (4):
John sells 100 shares of XYZ Corp at a loss.
A week later, his spouse, Jane, buys 50 shares of XYZ Corp.
Since they are married and living together:

John’s loss on the 100 shares he sold will be partially restricted.
Only the portion of the loss corresponding to 50 shares (the difference between what John sold and what Jane bought) will be available for immediate deduction.
The rest will be deferred and only allowed as a deduction when Jane eventually sells her shares.
These rules aim to prevent “loss harvesting” between sales and repurchases, even across spouses, by delaying the tax benefits until there is a genuine change in investment holdings.

***You said:***
what are sections 538 or 546 

***ChatGPT said:***
In the context of the Capital Gains Tax Acts, sections 538 and 546 refer to provisions in tax law that govern how capital losses can be applied to offset capital gains.

Section 538: This section generally covers the relief for capital losses. It outlines how individuals can use their capital losses to reduce their overall taxable gains. Losses on the sale of certain assets, like shares, can normally be deducted from capital gains, thereby reducing the amount of tax owed.

Section 546: This section provides additional rules around allowable losses for capital gains tax purposes. It defines specific circumstances or restrictions under which losses are considered allowable (or not allowable) and provides guidance on how to claim these losses against future gains.

In the context of the previous rule (about reacquiring shares), these sections would typically allow a loss to be offset against future gains, but the “bed and breakfasting” rule (Section 3) restricts this benefit temporarily if the shares are reacquired too soon after the sale.

</details>

## Notes

### Ringfenced losses

Not going to implement ringfence of 4 weeks

As outlined in example 2 here: https://www.revenue.ie/en/gains-gifts-and-inheritance/transfering-an-asset/selling-or-disposing-of-shares.aspx

```
Shares sold within four weeks of acquisition

Shares bought and sold within a four-week period cannot be offset against other gains.

You can only deduct the loss from a gain made on a subsequent disposal of same-class shares acquired within the four weeks.

    Example 2

    On 1 April 2017, both Jane and Kevin individually bought 3,000 ordinary shares in Abcee Ltd for €3,000.

    They both then sold their shares on 14 April 2017 for €2,000, making a loss of €1,000.

    Jane did not buy any more ordinary shares in Abcee Ltd within four weeks making the loss. She cannot set her loss against any gain she may make.

    Kevin bought more ordinary shares in Abcee Ltd on 21 April 2017. If Kevin makes a gain on the disposal of these shares in the future, he can deduct his loss of €1,000.
```

Reasons to not do this
- Not really applicable for my datasets (manually checked)
- Introducing this adds a lot of complexity with regards to how to handle the "pot" of losses that needs to be carried forward. It can be that one chooses to not use up this "pot" of losses in the year of filing for use later, or chose to use it for the year. Options of partial usage of the pot can also apply. And the same options need to be applied for every ticker as well. Complexity and control flow explodes, so at this point you definitely want an accountant.
