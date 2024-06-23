# Deej Plus (Wireless Controller with UDP through Wifi and custom buttons support)
Forked project from https://github.com/iamjackg/deej/releases/latest
and 
https://github.com/mohasali/deej-ESP32-Pro/releases/latest

All libraries are updated to current date.

deej is originaly an **open-source hardware volume mixer** for Windows and Linux PCs. It lets you use real-life sliders (like a DJ, but wait I am a DJ!) to **seamlessly control the volumes of different apps** (such as your music player, the game you're playing and your voice chat session) without having to stop what you're doing.

**In addition to the usual serial communication, this fork supports receiving fader data via UDP as well as custom buttons mapping** If you have an ESP8266 or ESP32-based microcontroller, this will allow you to build a wireless controller that will talk to your PC via Wi-Fi.

Unfortunately although the original Deej concept is very interesting there are several issues that sent me on this path of developping my own version :

-Firstly the original project, while absolutely brilliant, it is stuck about 3 years ago and most active users have simply forked the project instead of submitting pull requests. This resulted in a constellation of micro adjustments all over the Internet. I am trying to consolidate most interesting features into this repository or at least what I find suitable for a recent ESP32 implementation, which leads us to the second gripe.

-Originaly intended for newbies in the world of electronics, the project focused solely on Arduino as a client MCU. Fine and dandy when I started electronics in 2011 but nowadays for the price of a Nano you can get a complete state of the art SoC such as an ESP32. It makes no sense at all to still recommend an Arduino as a base for such a project except for one striking element. The ADC. Indeed Arduinos do have better embedded ADC compared to Espressif offerings, but there are ways around this. Parts are easily obtainable and one can easily use super dupper 20bits ADC for a penny if required. With my tests I realized that although ADC on ESP32 were at best mediocre out of the box, there are easy tricks to use them to their full potential, which can be good enough for 99% of Deej implementations. Those willing to make the extra miles will be able to use a dedicated ADC with I2C or SPI interface to the MCU.

-While the original project claims loud and clear that its first and foremost goal is ease of implementation (and thus still recommends Arduinos), the server side code base is written in Go langage, which contradicts the first assumption. The community of Python users is much more extended. Nevertheless I have learned to deal with it and indeed it is a good choice of langage and I understand it better now. But at the same time it makes code evolution much more involved than required for newbies.

-Thirdly the main reason why I started this fork is because I have been a computer infrastructure specialist for 35 years and a DJ for even longer. So I know my fair deal on hardware controllers. When I bumped into the Streamdesk trend at my work, I realized the limitations of this product (its owned by Corsair after all...). It is not developped with an open mind, it is centered on very narrow use cases and will not even run on Linux. In 2024 this is unacceptable, and even worse, the manufacturer will not even consider porting or opening its code base. For reference AMD and Nvidia have both opened their driver code base. Streamdesk don't have any fader support and their latest offerings only have a few rotary encoders. They are expensive and not suitable for my home studio where I need to adjust my sound system as much as I need to trigger various actions on my OS.

Overall I find Deej to be a great development base to build the perfect hardware controller in 2024, but there is much work to be done...

## Serial  support

Works the same as the stock definitions from the deej repository. The serial output will be sent to the defined serial bus such as /dev/ttyACM0 on Linux or COMx for Windows.

## UDP support

The configuration file for this fork has two extra options.

```yaml
# set this to "serial" to use deej with a board connected via a serial port
# set this to "udp" to listen for fader movements via a UDP network connection
controller_type: serial

# settings for the UDP connection
udp_port: 16990
```

If you set `controller_type` to `udp`, deej will start listening for UDP packets on the specified `udp_port` instead of opening the serial interface.

The deej protocol is very simple: each packet must consist of a series of numbers between 0 and 1023 for each slider, separated by a `|`. No newline is necessary at the end of each packet.

For example, if you have 5 faders, and the second and third one are currently at the midpoint, your packets would look like this:

```text
Slider|1023|42|591|42|42
```

This means that you can control deej with anything that can send UDP packets, including Wi-Fi enabled microcontrollers like the ESP8266 and ESP32, or scripts, or your home automation system, or anything you want!

This build brings custom mappable buttons support for OS interpretation allowing you to trigger any single key shortcut you want (I have not managed to implement multiple key shortcuts yet sorry...) or any binary or URL on your system.
You will thus find an extra definition in the config.yaml file under the block :

```yaml
# use keyboard. follwed by a keyboard key
# look at button.go for the list of keys
# you can open a file by giving the file location
# use / when giving a file location in both windows and linux
# buttons just run a cmd line so you can basically do anything with it.
# just like in sliders you can list commands for a button to execute

button_mapping:
  0: keyboard.lxmute
  1: keyboard.lxprevioussong
  2: keyboard.lxplaypause
  3: keyboard.lxnextsong
  4: strawberry
  5: google-chrome https://youtube.com/
  6: google-chrome https://gmail.com/
  7: mate-terminal
  8: google-chrome
```
Similarly to fader, the controller will output its button status under the string :

```text
Button|0|0|0|0|1
```
Then will send a release signal after a user definable duration :
```text
Button|0|0|0|0|0
```

Thus similar to a direct keyboard input.
Keep in mind that the current implementation of the keyboard events library does not directly account for multitple keys combinations such as SHIFT+F10 for instance. You may have to adapt your system shortcuts if dealing with exotic circumstances. It may sound simple but it seems like a big pile of complexity to implement at least with my level in Go langage.
On the long run a complete code revamping should ideally bring such a functionality.

### Building the controller

This repo only addresses code for ESP32 based controllers, but there is no reason it should not work on stock Arduino as long as you have the required hardware.
The basics are exactly the same as what is listed in the repo for the original project. In UDP mode, the main difference is that the final string should be sent via UDP instead of through the serial port. The controller sends both anyways.

The current version of the controller firmware and hardware are being actively developped in order to implement further functionnalities. More to come soon!

## Use case

This version currently runs on Linux Mint 21.3 with full Pipewire support and kernel 6.5.0-41-generic.

Not tested on Windows 11 yet.

## Feature and pull requests

If interested in the project feel free to send a pull request or a feature request. I will do my best to address them.

## License

deej is released under the [MIT license](./LICENSE).
