WSNet2 Lobby API
================

## Create Room

ednpoint
: POST /rooms

### エラーレスポンス
| 概要 | HTTP Status | gRPC Code | 発生箇所  | 備考 |
|------|-------------|-----------|-----------|------|
| レスポンスのmsgpackエンコード失敗 | InternalServerError | - | lobby/service/api.go: renderResponse() | - |
| ユーザ認証失敗 | Unauthorized | - | lobby/service/api.go: LobbyService.authUser() | - |
| リクエストbodyのmsgpackデコード失敗 | BadRequest | - | lobby/service.api.go: handleCreateRoom() | - |
| appIdのAppが無い | InternalServerError | - | lobby/room.go: RoomService.Create() | ユーザ認証失敗しているはずなので起こらない |
| gameサーバ取得失敗 | InternalServerError | - | lobby/game_cache.go: GameCache.Rand() | 生きているgameが見つからない |
| gRPC ClientをPoolから取得失敗 | InternalServerError | - | lobby/room.go: RoomService.Create() | - |
| gRPCタイムアウト | InternalServerError | DeadlineExceeded | lobby/room.go: RoomService.Create() | lobby側で設定したタイムアウト |
| appIdのAppが無い | InternalServerError | Internal | game/service/grpc.go: GameService.Create() | ユーザ認証失敗しているはずなので起こらない |
| context timeout | InternalServerError | DeadlineExceeded | game/repository.go: Repository.CreateRoom(), game/room.go: NewRoom() | game側で設定したタイムアウト |
| roomレコード作成失敗 | InternalServerError | Internal | game/repository.go: Repository.newRoomInfo() | - |
| room数上限 | ServiceUnavailable | ResourceExhausted | game/repository.go: Repository.CreateRoom() | - |
| DB Commit失敗 | InternalServerError | Internal | game/repository.go: Repository.CreateRoom() | - |
| {public,private}PropsのUnmarshal失敗| BadRequest | InvalidArgument | game/room.go: NewRoom() | - |
| Player PropsのUnmarshal失敗 | BadRequest | InvalidArgument | game/client.go: newClient() | - |


