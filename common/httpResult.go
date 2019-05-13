package common

//RequstResult all http request result data of the http request return to the client
type RequstResult struct {
	Code    ErrorCode   `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Host    string      `json:"host,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

//ErrorCode define code and string
type ErrorCode int

const (
	//Success 返回成功
	Success ErrorCode = iota + 10000
	//LocalNetworkError 本地网络错误，socket错误
	LocalNetworkError
	//CredentialError 认证失败
	CredentialError
	//ServiceNotFound 服务(docker)没有发现,可能不在运行
	ServiceNotFound
	//ServiceNotHealth 健康检查不成功
	ServiceNotHealth
	//PullImageError 获取docker iamge失败
	PullImageError
	//StopContainerError 停止容器失败
	StopContainerError
	//StartContainerError 运行容器失败
	StartContainerError
	//ParamsNotValide 参数错误
	ParamsNotValide
	//UnKnowServerError 服务器没有处理到的错误，比如服务器内部错误等
	UnKnowServerError
	//CannotMapToResult http 返回数据无法转为RequstResult
	CannotMapToResult
	//CannotGetHostIP 无法获取host ip
	CannotGetHostIP
	//StartServiceError 启动serveice失败
	StartServiceError
	//CheckServiceHealthError 服务的监控检查失败
	CheckServiceHealthError
	//GetServiceStatusError 获取服务器的状态失败
	GetServiceStatusError
	//CannotFoundAvalibleHost 运行服务，但是没有发现可用的host
	CannotFoundAvalibleHost
	//NotAllowedRuningOnHost 不允许在host运行
	NotAllowedRuningOnHost
	//AgentHaveNotRegistedOnProxy agent没有在proxy上注册
	AgentHaveNotRegistedOnProxy
)

func (me ErrorCode) String() string {
	switch me {
	case Success:
		return "Success"
	case ServiceNotFound:
		return "Service Not Found"
	case ServiceNotHealth:
		return "Service Not Health"
	case PullImageError:
		return "Pull Image Failt"
	case StopContainerError:
		return "Stop Container Failt"
	case LocalNetworkError:
		return "Local network error"
	case UnKnowServerError:
		return "Unkonw server error"
	case CannotMapToResult:
		return "Can not map to result"
	case CannotGetHostIP:
		return "Can not get the Host IP"
	case StartServiceError:
		return "Start server failt"
	case CheckServiceHealthError:
		return "Check server health failt"
	case GetServiceStatusError:
		return "Get server status failt"
	case CannotFoundAvalibleHost:
		return "Can not found the avalible host for the serivce"
	case NotAllowedRuningOnHost:
		return "Not allowed to running on host"
	case AgentHaveNotRegistedOnProxy:
		return "Agent Have not Registed on Proxy"
	default:
		return "Unknow"
	}
}

//NewResult http server put the result to the client
func NewResult(errorCode ErrorCode) *RequstResult {
	result := &RequstResult{Code: errorCode, Message: errorCode.String()}
	return result
}

//func NewHTTPResult
