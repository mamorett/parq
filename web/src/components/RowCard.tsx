import { useState } from 'react';
import { Card, Elevation, H5, Text, Tag, Button, Dialog, Classes } from '@blueprintjs/core';
import type { Row, Config } from '../types';
import { Thumbnail } from './Thumbnail';
import { getDownloadUrl, getThumbnailUrl, getFileUrl } from '../api';
import { formatDate } from '../utils';

export function RowCard({ row, schema }: { row: Row; schema: Config }) {
  const [isOpen, setIsOpen] = useState(false);
  const thumbnailCol = schema.thumbnail.column;

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const formatValue = (col: any, val: any) => {
    if (col.format === 'datetime') return formatDate(val);
    return String(val);
  };

  const FieldRow = ({ col, value }: { col: any, value: any }) => (
    <div className="field-row">
      <b style={{ color: 'var(--nord9)', minWidth: '120px' }}>{col.label}:</b>
      <span className="field-value">{formatValue(col, value)}</span>
      <Button
        icon="clipboard"
        minimal
        small
        className="copy-btn"
        onClick={(e) => { e.stopPropagation(); copyToClipboard(formatValue(col, value)); }}
      />
    </div>
  );

  return (
    <>
      <Card
        interactive
        elevation={Elevation.TWO}
        onClick={() => setIsOpen(true)}
        style={{ backgroundColor: 'var(--nord2)', color: 'var(--nord6)', padding: '1.25rem', width: '100%' }}
      >
        <div style={{ display: 'flex', gap: '2rem' }}>
          <div style={{ width: '180px', flexShrink: 0 }}>
            <Thumbnail index={row.index} column={thumbnailCol} />
          </div>
          <div style={{ flex: 1, overflow: 'hidden', textAlign: 'left' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1rem' }}>
              <H5 style={{ color: 'var(--nord8)', margin: 0, textAlign: 'left' }}>Row #{row.index}</H5>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <Button icon="maximize" minimal small onClick={(e) => { e.stopPropagation(); setIsOpen(true); }} />
                <Button icon="download" minimal small onClick={(e) => { e.stopPropagation(); window.open(getDownloadUrl(row.index, thumbnailCol)); }} />
              </div>
            </div>

            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.75rem', marginBottom: '1.25rem', justifyContent: 'flex-start' }}>
              {row.image_meta && (
                <>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem', backgroundColor: 'var(--nord1)', padding: '2px 8px', borderRadius: '4px' }}>
                    <Tag minimal intent="primary">{row.image_meta.width}x{row.image_meta.height}</Tag>
                    <Button icon="clipboard" minimal small className="copy-btn" onClick={(e) => { e.stopPropagation(); copyToClipboard(`${row.image_meta?.width}x${row.image_meta?.height}`); }} />
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem', backgroundColor: 'var(--nord1)', padding: '2px 8px', borderRadius: '4px' }}>
                    <Tag minimal intent="warning">{row.image_meta.aspect}</Tag>
                    <Button icon="clipboard" minimal small className="copy-btn" onClick={(e) => { e.stopPropagation(); copyToClipboard(row.image_meta?.aspect || ''); }} />
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem', backgroundColor: 'var(--nord1)', padding: '2px 8px', borderRadius: '4px' }}>
                    <Tag minimal intent="success">{row.image_meta.file_size_kb.toFixed(1)} KB</Tag>
                    <Button icon="clipboard" minimal small className="copy-btn" onClick={(e) => { e.stopPropagation(); copyToClipboard(row.image_meta?.file_size_kb.toFixed(1) + " KB"); }} />
                  </div>
                </>
              )}
            </div>

            <div style={{ display: 'flex', flexDirection: 'column' }}>
              {schema.columns.filter(c => !c.hidden).slice(0, 8).map(col => (
                <FieldRow key={col.name} col={col} value={row.columns[col.name]} />
              ))}
              {schema.columns.length > 8 && (
                <Text style={{ color: 'var(--nord3)', fontStyle: 'italic', marginTop: '0.5rem', textAlign: 'left' }}>+ {schema.columns.length - 8} more fields (click to expand)</Text>
              )}
            </div>
          </div>
        </div>
      </Card>

      <Dialog
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        title={`Row #${row.index} Details`}
        className={Classes.DARK}
        style={{ width: '95%', maxWidth: '1200px', backgroundColor: 'var(--nord0)' }}
      >
        <div className={Classes.DIALOG_BODY} style={{ color: 'var(--nord6)', textAlign: 'left' }}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '2rem' }}>
            <div style={{ textAlign: 'left' }}>
              <a href={getFileUrl(row.index, thumbnailCol)} target="_blank" rel="noreferrer">
                <img
                  src={getFileUrl(row.index, thumbnailCol)}
                  alt="Full Res"
                  className="detail-dialog-img"
                />
              </a>
              <H5 style={{ color: 'var(--nord8)', borderBottom: '1px solid var(--nord2)', paddingBottom: '0.5rem', marginTop: '1.5rem', textAlign: 'left' }}>Image Properties</H5>
              {row.image_meta ? (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', backgroundColor: 'var(--nord1)', padding: '1rem', borderRadius: '4px' }}>
                  <div className="field-row">
                    <b style={{ minWidth: '100px' }}>Dimensions:</b> {row.image_meta.width} x {row.image_meta.height} px
                    <Button icon="clipboard" minimal small className="copy-btn" onClick={() => copyToClipboard(`${row.image_meta?.width}x${row.image_meta?.height}`)} />
                  </div>
                  <div className="field-row">
                    <b style={{ minWidth: '100px' }}>Aspect:</b> {row.image_meta.aspect}
                    <Button icon="clipboard" minimal small className="copy-btn" onClick={() => copyToClipboard(row.image_meta?.aspect || '')} />
                  </div>
                  <div className="field-row">
                    <b style={{ minWidth: '100px' }}>Size:</b> {row.image_meta.file_size_kb.toFixed(1)} KB
                    <Button icon="clipboard" minimal small className="copy-btn" onClick={() => copyToClipboard(row.image_meta?.file_size_kb.toFixed(1) + " KB")} />
                  </div>
                </div>
              ) : (
                <Text style={{ color: 'var(--nord3)', textAlign: 'left' }}>No image properties available.</Text>
              )}
            </div>
            
            <div style={{ overflow: 'auto' }}>
              <H5 style={{ color: 'var(--nord8)', borderBottom: '1px solid var(--nord2)', paddingBottom: '0.5rem', textAlign: 'left' }}>Metadata</H5>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                {schema.columns.map(col => (
                  <div key={col.name} style={{ marginBottom: '0.75rem' }}>
                    <div style={{ color: 'var(--nord9)', fontWeight: 'bold', fontSize: '0.85rem', marginBottom: '0.2rem', textAlign: 'left' }}>{col.label}</div>
                    <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center', backgroundColor: 'var(--nord1)', padding: '0.5rem', borderRadius: '4px' }}>
                      <div style={{ color: 'var(--nord4)', wordBreak: 'break-all', flex: 1, textAlign: 'left' }}>
                        {formatValue(col, row.columns[col.name])}
                      </div>
                      <Button
                        icon="clipboard"
                        minimal
                        small
                        onClick={() => copyToClipboard(formatValue(col, row.columns[col.name]))}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
        <div className={Classes.DIALOG_FOOTER}>
          <div className={Classes.DIALOG_FOOTER_ACTIONS}>
            <Button onClick={() => setIsOpen(false)}>Close</Button>
          </div>
        </div>
      </Dialog>
    </>
  );
}
