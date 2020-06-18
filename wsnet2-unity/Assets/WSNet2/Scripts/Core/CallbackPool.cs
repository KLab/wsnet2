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
    /// </remarks>
    public class CallbackPool
    {
        // TODO: delegateだと途中で例外投げるやつがいたとき困るのでQueueにしたい
        Action current;
        Action next;
        object processLock = new object();

        public void Add(Action callback)
        {
            lock(this)
            {
                current += callback;
            }
        }

        public void Clear()
        {
            lock(this)
            {
                current = null;
            }
        }

        public void Process()
        {
            lock(processLock){
                lock(this)
                {
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
