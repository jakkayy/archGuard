# Record the README demo GIF with https://github.com/charmbracelet/vhs:
#   vhs docs/demo.tape
Output docs/demo.gif
Set FontSize 16
Set Width 1100
Set Height 650
Set Theme "Catppuccin Mocha"

Hide
Type "cd $(mktemp -d) && git init -q && echo '# demo' > README.md && clear"
Enter
Show

Type "printf 'package main\n\nconst awsKey = \"AKIA%s\"\n' IOSFODNN7EXAMPLE > config.go"
Enter
Sleep 500ms
Type "archguard init -y"
Enter
Sleep 1s
Type "archguard scan"
Enter
Sleep 3s
Type "echo 'bin/' > .gitignore && sed -i '/awsKey/d' config.go"
Enter
Sleep 500ms
Type "archguard scan"
Enter
Sleep 3s
