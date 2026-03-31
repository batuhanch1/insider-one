package utils

const LogLevel_Error = "Error"
const LogLevel_Information = "Information"
const LogStatus_Executing = "Executing"
const LogStatus_Executed = "Executed"
const LogStatus_StartRequest = "Started Request"
const LogStatus_FinishRequest = "Finish Request"
const LogStatus_DbQueryStart = "Db Query Started"
const LogStatus_DbQueryFinish = "Db Query Finish"
const LogStatus_MessageWritedQueue = "Message Writed To Queue"
const LogStatus_MessageReadedQueue = "Message Readed From Queue"

const Layout_TimeWithNano = "2006-01-02T15:04:05.999999999Z07:00"
const Layout_Time = "2006-01-02 15:04:05"

const Header_CorrelationID = "IO-CorrelationID"
const Header_ExternalRequestStartTime = "IO-ExternalRequestStartTime"
const Header_InternalRequestStartTime = "IO-InternalRequestStartTime"
const Header_QueryStartTime = "IO-QueryStartTime"
