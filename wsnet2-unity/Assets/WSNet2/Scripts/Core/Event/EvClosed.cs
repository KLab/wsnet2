namespace WSNet2
{
    public class EvClosed : Event
    {
        public string Description { get; private set; }

        public EvClosed(string description) : base(EvType.Closed, null)
        {
            Description = description;
        }
    }
}
