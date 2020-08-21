using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;

namespace Sample
{
    /// <summary>
    /// InputField のデフォルト文字を設定するコンポーネント
    /// </summary>
    public class InputFieldScript : MonoBehaviour
    {
        public string deafultText;

        void Start()
        {
            if (this.deafultText != null)
            {
                var inputField = GetComponent<InputField>();
                inputField.text = deafultText;
            }
        }

        void Update()
        {

        }
    }
}