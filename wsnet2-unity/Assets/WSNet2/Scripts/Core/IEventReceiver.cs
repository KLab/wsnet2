using System;

namespace WSNet2.Core
{
    /// <summary>
    ///   イベントを受け取るObjectのInterface
    /// </summary>
    /// <remarks>
    ///   <para>
    ///     これらのメソッドが呼ばれるときは、既にRoomに情報反映済み。
    ///     プロパティなど変更前の値が必要な場合はアプリ側で保持しておくこと。
    ///   </para>
    /// </remarks>
    public interface IEventReceiver
    {
        /// <summary>何らかのエラー</summary>
        /// <remarks>
        ///   <para>
        ///     接続維持できなくなった状態。
        ///     自動再接続できる場合は呼ばれず勝手に再接続する。
        ///   </para>
        /// </remarks>
        /// <param name="e">原因の例外</param>
        void OnError(Exception e);

        /// <summary>
        ///   自分自身の入室通知
        /// </summary>
        /// <param name="me">自分自身</param>
        void OnJoined(Player me);

        /// <summary>
        ///   新規プレイヤーの入室
        /// </summary>
        /// <param name="player">入室したプレイヤー</param>
        void OnOtherPlayerJoined(Player player);

        /// <summary>
        ///   汎用メッセージ
        /// </summary>
        /// <param name="ev">イベント</param>
        /// <remarks>
        ///   <para>
        ///     メッセージの中身を取得するにはev.GetBody()を呼び出す。
        ///     このメソッド終了後にevは再利用に回され利用できなくなるため、
        ///     かならずこの中でGetBody()しておくこと。
        ///     GetBody()で取得したオブジェクトはrecycle引数に渡したものか新規確保したものなので
        ///     このメソッド終了後も利用できる。
        ///   </para>
        /// </remarks>
        void OnMessage(EvMessage ev);
    }
}
