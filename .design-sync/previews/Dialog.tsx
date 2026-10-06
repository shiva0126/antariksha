import { Button, Dialog } from '@astrisk/ds';


const ConfirmStory = () => (
  <Dialog open title="Delete this chart?" onClose={() => {}}>
    <div style={{ display: 'grid', gap: 16 }}>
      <p style={{ margin: 0 }}>"Ananya Rao" will be removed from your saved charts. This cannot be undone.</p>
      <div className="dialog-actions"><Button variant="ghost">Cancel</Button><Button variant="danger">Delete</Button></div>
    </div>
  </Dialog>
);

export const Confirm = ConfirmStory;
