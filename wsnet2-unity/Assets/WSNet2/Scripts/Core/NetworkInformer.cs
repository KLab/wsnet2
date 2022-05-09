using System;
using System.Diagnostics;
using System.Net;

namespace WSNet2
{
    /// <summary>
    /// ネットワークアクセスの情報通知クラス
    /// </summary>
    /// WSNet2内部で通信が発生したことを通知する仕組みを提供する。
    /// NetworkInformer.SetCallback() で通知先を設定することで有効化。
    /// Debugビルドでのみ有効。
    public class NetworkInformer
    {
        /// <summary>
        ///   送受信データ情報
        /// </summary>
        public abstract class Info
        {
            /// <summary>送信データ本体のサイズ</summary>
            /// <remarks>各種ヘッダ情報は含まない（HTTP、WebSocket、TCP、IPなど）</remarks>
            public int BodySize;
        }

        /// <summary>
        ///   Lobbyへの送信データ情報（HTTP）
        /// </summary>
        public class LobbySendInfo : Info
        {
            /// <summary>リクエストURL</summary>
            public string URL;
        }

        /// <summary>
        ///   Lobbyからの受信データ情報（HTTP）
        /// </summary>
        public class LobbyReceiveInfo : Info
        {
            /// <summary>リクエストURL</summary>
            public string URL;

            /// <summary>レスポンスのStatusCode</summary>
            public HttpStatusCode StatusCode;
        }

        /// <summary>
        ///   Roomへの接続リクエスト送信（HTTP）
        /// </summary>
        public class RoomConnectRequestInfo : Info
        {
            /// <summary>リクエストURL</summary>
            public string URL;

            /// <summary>部屋ID</summary>
            public string RoomID;
        }

        /// <summary>
        ///   Roomへの送信データ情報（WebSocket）
        /// </summary>
        public class RoomSendInfo : Info
        {
            /// <summary>部屋ID</summary>
            public string RoomID;

            /// <summary>メッセージ種別</summary>
            public MsgType MsgType;

            /// <summary>RPC情報</summary>
            /// <remarks>MsgType.Bloadcast/Target/ToMasterのときのみ有効</remarks>
            public RPCInfo RPCInfo;
        }

        /// <summary>
        ///   Roomからの受信データ情報（WebSocket）
        /// </summary>
        public class RoomReceiveInfo : Info
        {
            /// <summary>部屋ID</summary>
            public string RoomID;

            /// <summary>イベント種別</summary>
            public EvType EvType;

            /// <summary>RPC情報</summary>
            /// <remarks>EvType.Messageのときのみ有効</remarks>
            public RPCInfo RPCInfo;
        }

        /// <summary>
        ///   RPC情報
        /// </summary>
        public class RPCInfo
        {
            /// <summary>Roomに登録されたRPC ID</summary>
            public int ID;

            /// <summary>メソッド名</summary>
            public string MethodName;
        }

#if DEBUG
        static Action<Info> callback = null;
        static object callbackLock = new object();
#endif

        /// <summary>
        ///   通知を受け取るCallbackを設定
        /// </summary>
        /// <param name="callback">通知を受け取るCallback</param>
        /// <remarks>
        ///   callbackをnullにすることで無効化できる
        /// </remarks>
        [Conditional("DEBUG")]
        public static void SetCallback(Action<Info> callback)
        {
#if DEBUG
            lock (callbackLock)
            {
                NetworkInformer.callback = callback;
            }
#endif
        }

        /// <summary>
        ///   Lobbyへの送信発生（内部用）
        /// </summary>
        [Conditional("DEBUG")]
        public static void OnLobbySend(string url, byte[] body)
        {
#if DEBUG
            lock (callbackLock)
            {
                callback?.Invoke(
                    new LobbySendInfo()
                    {
                        URL = url,
                        BodySize = body.Length,
                    });
            }
#endif
        }

        /// <summary>
        ///   Lobbyからの受信発生（内部用）
        /// </summary>
        [Conditional("DEBUG")]
        public static void OnLobbyReceive(string url, HttpStatusCode code, byte[] body)
        {
#if DEBUG
            lock (callbackLock)
            {
                callback?.Invoke(
                    new LobbyReceiveInfo()
                    {
                        URL = url,
                        StatusCode = code,
                        BodySize = body.Length,
                    });
            }
#endif
        }

        /// <summary>
        ///   Roomへの接続リクエスト発生（内部用）
        /// </summary>
        [Conditional("DEBUG")]
        public static void OnRoomConnectRequest(Room room, string url)
        {
#if DEBUG
            lock (callbackLock)
            {
                callback?.Invoke(
                    new RoomConnectRequestInfo()
                    {
                        URL = url,
                        RoomID = room.Id,
                        BodySize = 0,
                    });
            }
#endif
        }

        /// <summary>
        ///   Roomへの送信発生（内部用）
        /// </summary>
        [Conditional("DEBUG")]
        public static void OnRoomSend(Room room, ArraySegment<byte> payload)
        {
#if DEBUG
            lock (callbackLock)
            {
                if (callback == null)
                {
                    return;
                }

                var (msgType, rpcId) = MsgPool.ParsePayload(payload);
                RPCInfo rpcInfo = null;
                if (rpcId.HasValue)
                {
                    rpcInfo = new RPCInfo()
                    {
                        ID = rpcId.Value,
                        MethodName = room.MethodName(rpcId.Value),
                    };
                }

                callback(
                    new RoomSendInfo()
                    {
                        BodySize = payload.Count,
                        RoomID = room.Id,
                        MsgType = msgType,
                        RPCInfo = rpcInfo,
                    });
            }
#endif
        }

        /// <summary>
        ///   Roomからの受信発生（内部用）
        /// </summary>
        [Conditional("DEBUG")]
        public static void OnRoomReceive(Room room, int bodySize, Event ev)
        {
#if DEBUG
            lock (callbackLock)
            {
                if (callback == null)
                {
                    return;
                }

                RPCInfo rpcInfo = null;
                var evrpc = ev as EvRPC;
                if (evrpc != null)
                {
                    rpcInfo = new RPCInfo
                    {
                        ID = evrpc.RpcID,
                        MethodName = room.MethodName(evrpc.RpcID),
                    };
                }

                callback(
                    new RoomReceiveInfo()
                    {
                        BodySize = bodySize,
                        RoomID = room.Id,
                        EvType = ev.Type,
                        RPCInfo = rpcInfo,
                    });
            }
#endif
        }
    }
}
