using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;

namespace Sample
{
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