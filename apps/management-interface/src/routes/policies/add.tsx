import { createFileRoute, useNavigate } from '@tanstack/react-router'
import {Description, Field, FieldGroup, Fieldset, Label} from "@/components/fieldset.tsx";
import {Input} from "@/components/input.tsx";
import {Select} from "@/components/select.tsx";
import {Textarea} from "@/components/textarea.tsx";
import {Heading} from "@/components/heading.tsx";
import {Button} from "@/components/button.tsx";
import { useState } from 'react';
import { useAddPolicy } from '@/services/policies';

export const Route = createFileRoute('/policies/add')({
  component: AddPolicyComponent,
})

function AddPolicyComponent() {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    policy_name: '',
    language: 'cedar',
    data: '',
    rvvaId: '',
    description: '',
    tags: ['']
  });
  const [error, setError] = useState<string | null>(null);

  const addPolicyMutation = useAddPolicy();

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!formData.description || !formData.language || !formData.data) {
      setError('All fields are required');
      return;
    }

    const id = (Math.floor(Math.random() * 9000) + 1000).toString()
    try {
      await addPolicyMutation.mutateAsync({
        language: formData.language.toLowerCase(),
        id: id,
        policy: {
          id: id,
          language: formData.language.toLowerCase(),
          data: formData.data,
          rvvaId: formData.rvvaId,
          description: formData.description,
        }
      });

      // Redirect to policies list on success
      await navigate({to: '/policies'});
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add policy');
    }
  };

  return (
      <>
        <form onSubmit={handleSubmit}>
          <Heading>Add policy</Heading>
          {error && (
            <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
              {error}
            </div>
          )}
          <Fieldset>
            <FieldGroup>
              <Field>
                <Label>Policy description</Label>
                <Input 
                  name="description"
                  value={formData.description}
                  onChange={handleChange} 
                  required 
                />
              </Field>
              <Field>
                <Label>Language</Label>
                <Select 
                  name="language" 
                  value={formData.language}
                  onChange={handleChange}
                >
                  <option value="cedar">Cedar</option>
                  <option value="rego">Rego</option>
                  <option value="cerbos">Cerbos</option>
                  <option value="openfga">OpenFGA</option>
                </Select>
                <Description>The language of the policy that is added.</Description>
              </Field>
              <Field>
                <Label>rvva ID</Label>
                <Input
                    name="rvvaId"
                    value={formData.rvvaId}
                    onChange={handleChange}
                />
              </Field>
              <Field>
                <Label>Policy code</Label>
                <Textarea 
                  name="data" 
                  rows={10} 
                  value={formData.data} 
                  onChange={handleChange}
                  required
                />
              </Field>
            </FieldGroup>
            <FieldGroup>
              <Fieldset className={"flex justify-between"}>
                <Button type="button" href="/policies/" color={"zinc"}>Discard</Button>
                <Button 
                  type="submit" 
                  color={"emerald"} 
                  disabled={addPolicyMutation.isPending}
                >
                  {addPolicyMutation.isPending ? 'Adding...' : 'Add policy'}
                </Button>
              </Fieldset>
            </FieldGroup>
          </Fieldset>
        </form>
      </>
  )
}
