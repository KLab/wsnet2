using System;

namespace WSNet2.Core
{
    /// <summary>
    ///   Callbackを溜めておいて後から実行できるようにするやつ.
    /// </summary>
    /// <remarks>
    ///   <para>
    ///     Unityではcallbackをメインスレッドで動かしたいので溜めておいて、
    ///     メインスレッドでProcess()を呼び出すようにする。
    ///   </para>
    ///   <para>
    ///     DotNetの場合にも適当なスレッドで定期的にProcess()を呼び出す必要がある。
    ///   </para>
    /// </remarks>
    public class CallbackPool
    {
        // TODO: delegateだと途中で例外投げるやつがいたとき困るのでQueueにしたい
        Action current;
        Action next;
        object processLock = new object();

        /// <summary>
        ///   callbackをpoolに追加
        /// </summary>
        public void Add(Action callback)
        {
            lock(this)
            {
                current += callback;
            }
        }

        /// <summary>
        ///   Callbackを追加された順に実行する
        /// </summary>
        public void Process()
        {
            lock(processLock){
                lock(this)
                {
                    // 実行中も別スレッドからAddできるようにプールを入れ替える
                    var tmp = current;
                    current = next;
                    next = tmp;
                }

                if (next != null)
                {
                    next();
                    next = null;
                }
            }
        }
    }
}
