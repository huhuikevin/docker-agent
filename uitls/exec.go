package uitls
import (
	"fmt"
	"bufio"
	"io"
	"os/exec"
)
func execCommand(commandName string, params []string) ([]string){
	var results = make([]string, 0)
    //函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
    cmd := exec.Command(commandName, params...)

    //显示运行的命令
    fmt.Println(cmd.Args)
    //StdoutPipe方法返回一个在命令Start后与命令标准输出关联的管道。Wait方法获知命令结束后会关闭这个管道，一般不需要显式的关闭该管道。
    stdout, err := cmd.StdoutPipe()

    if err != nil {
        fmt.Println(err)
        return results
    }

	serr := cmd.Start()
	if serr != nil {
		fmt.Println(serr)
		return results
	}
    //创建一个流来读取管道内内容，这里逻辑是通过一行一行的读取的
    reader := bufio.NewReader(stdout)

    //实时循环读取输出流中的一行内容
    for {
        line, err2 := reader.ReadString('\n')
        if err2 != nil || io.EOF == err2 {
            break
        }
		fmt.Println(line)
		results = append(results, line)
    }

    //阻塞直到该命令执行完成，该命令必须是被Start方法开始执行的
	cmd.Wait()
	fmt.Println(results)
	return results
}

